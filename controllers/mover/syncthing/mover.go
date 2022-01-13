/*
Copyright 2021 The VolSync authors.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package syncthing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	snapv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backube/volsync/api/v1alpha1"
	"github.com/backube/volsync/controllers"
	"github.com/backube/volsync/controllers/mover"
	"github.com/backube/volsync/controllers/utils"
)

// const (
// 	mountPath      = "/data"
// 	dataVolumeName = "data"
// )

// Mover is the reconciliation logic for the Restic-based data mover.
type Mover struct {
	client client.Client
	logger logr.Logger
	owner  metav1.Object
	// vh          *volumehandler.VolumeHandler
	isSource    bool
	paused      bool
	dataPVCName *string
	nodeList    []*v1alpha1.SyncthingNode
	// Source-only fields
	sourceStatus *v1alpha1.ReplicationSourceSyncthingStatus
	// Destination-only fields
	destinationStatus *v1alpha1.ReplicationDestinationSyncthingStatus
}

var _ mover.Mover = &Mover{}

// All object types that are temporary/per-iteration should be listed here. The
// individual objects to be cleaned up must also be marked.
var cleanupTypes = []client.Object{
	&corev1.PersistentVolumeClaim{},
	&snapv1.VolumeSnapshot{},
	&batchv1.Job{},
}

func (m *Mover) Name() string { return "syncthing" }

// We need the following resources available to us in the cluster:
// - PVC for syncthing-config
// - PVC that needs to be synced
// - Secret for the syncthing-apikey
// - Job/Pod running the syncthing mover image
// - Service exposing the syncthing REST API for us to make requests to
func (m *Mover) Synchronize(ctx context.Context) (mover.Result, error) {
	// ensure the data pvc exists
	if _, err := m.ensureDataPVC(ctx); err != nil {
		return mover.InProgress(), err
	}

	// create PVC for config data
	if _, err := m.ensureConfigPVC(ctx); err != nil {
		return mover.InProgress(), err
	}

	// ensure the secret exists
	if _, err := m.ensureSecretAPIKey(ctx); err != nil {
		return mover.InProgress(), err
	}

	// ensure the job exists
	if _, err := m.ensureJob(ctx); err != nil {
		return mover.InProgress(), err
	}

	// create the service for the syncthing REST API
	if _, err := m.ensureService(ctx); err != nil {
		return mover.InProgress(), err
	}

	// get the API key from the syncthing-apikey secret
	m.logger.Info("Getting API key")
	apiKey, err := m.getAPIKey(ctx)
	if err != nil {
		return mover.InProgress(), err
	}

	headers := map[string]string{
		"X-API-Key": apiKey,
	}

	_, err = controllers.JSONRequest("https://127.0.0.1:8384/rest/config", "GET", headers, nil, nil)
	if err != nil {
		return mover.InProgress(), err
	}
	// hello world example
	// send an API request to the service exposing syncthing's API
	// k8sClient.

	// On the source, just signal completion
	return mover.Complete(), nil
}

func (m *Mover) ensureConfigPVC(ctx context.Context) (*corev1.PersistentVolumeClaim, error) {
	/*
		# create the PVC syncthing-config with a small footprint
		apiVersion: v1
		kind: PersistentVolumeClaim
		metadata:
			name: syncthing-config
		spec:
			accessModes:
			- ReadWriteOnce
			resources:
				requests:
					storage: 1Gi
	*/
	configPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "syncthing-config",
			Namespace: m.owner.GetNamespace(),
		},
	}
	if err := m.client.Get(ctx, client.ObjectKeyFromObject(configPVC), configPVC); err == nil {
		// pvc already exists
		m.logger.Info("PVC already exists:  " + configPVC.Name)
		return configPVC, nil
	}

	// otherwise, create the PVC
	configPVC = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "syncthing-config",
			Namespace: m.owner.GetNamespace(),
			Labels: map[string]string{
				"app": "syncthing",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}
	if err := m.client.Create(ctx, configPVC); err != nil {
		return nil, err
	}
	m.logger.Info("Created PVC", configPVC.Name, configPVC)
	return configPVC, nil
}

func (m *Mover) ensureDataPVC(ctx context.Context) (*corev1.PersistentVolumeClaim, error) {
	// check if the data PVC exists, error if it doesn't
	fmt.Printf("Checking for PVC %s\n", *m.dataPVCName)
	dataPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *m.dataPVCName,
			Namespace: m.owner.GetNamespace(),
			Labels: map[string]string{
				"app": "syncthing",
			},
		},
	}
	if err := m.client.Get(ctx, client.ObjectKeyFromObject(dataPVC), dataPVC); err != nil {
		// pvc doesn't exist
		return nil, err
	}
	return dataPVC, nil
}

func (m *Mover) ensureSecretAPIKey(ctx context.Context) (*corev1.Secret, error) {
	/*
		The secret is in the following format:
		apiVersion: v1
		kind: Secret
		metadata:
			name: st-apikey
		type: Opaque
		data:
			apiKey: 'cGFzc3dvcmQxMjM='

	*/
	// check if the secret exists, error if it doesn't
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "syncthing-apikey",
			Namespace: m.owner.GetNamespace(),
			Labels: map[string]string{
				"app": "syncthing",
			},
		},
	}
	err := m.client.Get(ctx, client.ObjectKeyFromObject(secret), secret)

	if err != nil {
		// need to create the secret
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "syncthing-apikey",
				Namespace: m.owner.GetNamespace(),
				Labels: map[string]string{
					"app": "syncthing",
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				// base64 encode an empty string
				"apikey": []byte("password123"),
			},
		}
		if err := m.client.Create(ctx, secret); err != nil {
			// error creating secret
			m.logger.Error(err, "Error creating secret")
			return nil, err
		}
		m.logger.Info("Created secret", secret.Name, secret)
	}
	return secret, nil
}

//nolint:funlen
func (m *Mover) ensureJob(ctx context.Context) (*batchv1.Job, error) {
	/*
				Sample of the job as a yaml
				apiVersion: batch/v1
		kind: Job
		metadata:
		  name: syncthing
		spec:
		  ttlSecondsAfterFinished: 100
		  template:
		    spec:
		      restartPolicy: Never
		      containers:
		      - name: syncthing
		        image: "quay.io/backube/volsync-mover-syncthing"
		        command: ["/entry.sh"]
		        args:
		        - run
		        env:
		        - name: SYNCTHING_CONFIG_DIR
		          value: "/config"
		        - name: SYNCTHING_DATA_DIR
		          value: "/data"
		        - name: STGUIAPIKEY
		          valueFrom:
		            secretKeyRef:
		              name: syncthing-apikey
		              key: apikey
		        imagePullPolicy: Always
		        ports:
		        - containerPort: 8384
		        - containerPort: 22000
		        volumeMounts:
		        - name: syncthing-config
		          mountPath: /config
		        - name: synced-volume  # hook this up with whatever PVC you want to sync
		          mountPath: /data
		        resources:
		          limits:
		            cpu: 100m
		            memory: 1Gi
		      volumes:
		      - name: syncthing-config
		        persistentVolumeClaim:
		          claimName: syncthing-config
		      # the volume we want to sync
		      - name: synced-volume
		        persistentVolumeClaim:
		          # enter the PVC name here
		          claimName: dokuwiki-pvc
	*/
	// return successfully if the job exists, try to create it otherwise
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "syncthing",
			Namespace: m.owner.GetNamespace(),
			Labels: map[string]string{
				"app": "syncthing",
			},
		},
	}
	err := m.client.Get(ctx, client.ObjectKeyFromObject(job), job)
	if err == nil {
		// job already exists
		m.logger.Info("Job already exists: " + job.Name)
		return job, nil
	}
	if !errors.IsNotFound(err) {
		// something about the job is broken
		m.logger.Error(err, "Error getting job")
		return nil, err
	}

	var ttlSecondsAfterFinished int32 = 100

	// job doesn't exist, create it
	job = &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "syncthing",
			Namespace: m.owner.GetNamespace(),
			Labels: map[string]string{
				"app": "syncthing",
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "syncthing",
							Image: "quay.io/backube/volsync-mover-syncthing",
							Command: []string{
								"/entry.sh",
							},
							Args: []string{
								"run",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SYNCTHING_CONFIG_DIR",
									Value: "/config",
								},
								{
									Name:  "SYNCTHING_DATA_DIR",
									Value: "/data",
								},
								{
									Name: "STGUIAPIKEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "syncthing-apikey",
											},
											Key: "apikey",
										},
									},
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8384,
								},
								{
									ContainerPort: 22000,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "syncthing-config",
									MountPath: "/config",
								},
								{
									Name:      "synced-volume",
									MountPath: "/data",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "syncthing-config",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "syncthing-config",
								},
							},
						},
						{
							Name: "synced-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dokuwiki-pvc",
								},
							},
						},
					},
				},
			},
		},
	}

	// pass the object onto the k8s api
	err = m.client.Create(ctx, job)
	return job, err
}

func (m *Mover) ensureService(ctx context.Context) (*corev1.Service, error) {
	/* Service is in the following form:
	apiVersion: v1
	kind: Service
	metadata:
		name: syncthing-svc
	spec:
		selector:
			app: syncthing
		ports:
		- port: 8384
			targetPort: 8384
			protocol: TCP
	*/
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "syncthing",
			Namespace: m.owner.GetNamespace(),
			Labels: map[string]string{
				"app": "syncthing",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "syncthing",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8384,
					TargetPort: intstr.FromInt(8384),
					Protocol:   "TCP",
				},
			},
		},
	}
	err := m.client.Get(ctx, client.ObjectKeyFromObject(service), service)
	if err == nil {
		// service already exists
		m.logger.Info("service already exists", "service", service.Name)
		return service, nil
	}

	if err := m.client.Create(ctx, service); err != nil {
		m.logger.Error(err, "error creating the service")
		return nil, err
	}
	return service, nil
}

func (m *Mover) Cleanup(ctx context.Context) (mover.Result, error) {
	err := utils.CleanupObjects(ctx, m.client, m.logger, m.owner, cleanupTypes)
	if err != nil {
		return mover.InProgress(), err
	}
	return mover.Complete(), nil
}

// get the API key
func (m *Mover) getAPIKey(ctx context.Context) (string, error) {
	// get the syncthing-apikey secret
	secret := &corev1.Secret{}
	err := m.client.Get(ctx, client.ObjectKey{Name: "syncthing-apikey", Namespace: m.owner.GetNamespace()}, secret)
	if err != nil {
		return "", err
	}
	return string(secret.Data["apikey"]), nil
}
