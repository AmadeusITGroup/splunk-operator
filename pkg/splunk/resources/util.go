package resources

import (
	"fmt"
	"git.splunk.com/splunk-operator/pkg/splunk/enterprise"
	"git.splunk.com/splunk-operator/pkg/splunk/spark"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"log"
	"context"
	"git.splunk.com/splunk-operator/pkg/apis/enterprise/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)


func CreateResource(client client.Client, obj runtime.Object) error {
	err := client.Create(context.TODO(), obj)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Printf("Failed to create object : %v", err)
		return err
	}
	return nil
}


func AddOwnerRefToObject(object metav1.Object, reference metav1.OwnerReference) {
	object.SetOwnerReferences(append(object.GetOwnerReferences(), reference))
}


func AsOwner(instance *v1alpha1.SplunkEnterprise) metav1.OwnerReference {
	trueVar := true

	return metav1.OwnerReference{
		APIVersion: instance.TypeMeta.APIVersion,
		Kind:       instance.TypeMeta.Kind,
		Name:       instance.Name,
		UID:        instance.UID,
		Controller: &trueVar,
	}
}


func AddConfigMapVolumeToPodTemplate(podTemplateSpec *v1.PodTemplateSpec, volumeName string, configMapName string, mountLocation string) {
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	})
	for idx, _ := range podTemplateSpec.Spec.Containers {
		podTemplateSpec.Spec.Containers[idx].VolumeMounts = append(podTemplateSpec.Spec.Containers[idx].VolumeMounts, v1.VolumeMount{
			Name: volumeName,
			MountPath: mountLocation,
		})
	}
}


func AddLicenseVolumeToPodTemplate(podTemplateSpec *v1.PodTemplateSpec, volumeName string, source *v1.VolumeSource, mountLocation string) {
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, v1.Volume{
		Name: volumeName,
		VolumeSource: *source,
	})
	for idx, _ := range podTemplateSpec.Spec.Containers {
		podTemplateSpec.Spec.Containers[idx].VolumeMounts = append(podTemplateSpec.Spec.Containers[idx].VolumeMounts, v1.VolumeMount{
			Name: volumeName,
			MountPath: mountLocation,
		})
	}
}


func AddDFCVolumeMounts(containerSpec *v1.Container) {
	containerSpec.VolumeMounts = append(containerSpec.VolumeMounts, v1.VolumeMount{
		Name:      "mnt-jdk",
		MountPath: "/mnt/jdk",
	})
	containerSpec.VolumeMounts = append(containerSpec.VolumeMounts, v1.VolumeMount{
		Name:      "mnt-spark",
		MountPath: "/mnt/spark",
	})
}


func AddDFCToPodTemplate(podTemplateSpec *v1.PodTemplateSpec, instance *v1alpha1.SplunkEnterprise) error {
	requirements, err := GetSparkRequirements(instance)
	if err != nil {
		return err
	}

	containerSpec := corev1.Container {
		Image: spark.GetSparkImage(instance),
		ImagePullPolicy: corev1.PullPolicy(instance.Spec.Config.ImagePullPolicy),
		Name: "init",
		Resources: requirements,
		Command: []string{ "bash", "-c", "cp -r /opt/jdk /mnt && cp -r /opt/spark /mnt" },
	}

	AddDFCVolumeMounts(&containerSpec)
	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, containerSpec)

	emptyVolumeSource := corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}

	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
		Name: "mnt-jdk",
		VolumeSource: emptyVolumeSource,
	})
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
		Name: "mnt-spark",
		VolumeSource: emptyVolumeSource,
	})

	for idx, _ := range podTemplateSpec.Spec.Containers {
		AddDFCVolumeMounts(&podTemplateSpec.Spec.Containers[idx])
	}

	return nil
}


func UpdatePodTemplateWithConfig(podTemplateSpec *v1.PodTemplateSpec, cr *v1alpha1.SplunkEnterprise) error {
	dnsConfigSearches := enterprise.GetSplunkDNSConfiguration(cr)
	if dnsConfigSearches != nil {
		podTemplateSpec.Spec.DNSPolicy = corev1.DNSClusterFirst
		podTemplateSpec.Spec.DNSConfig = &corev1.PodDNSConfig{
			Searches: dnsConfigSearches,
		}
	}

	if cr.Spec.Config.SchedulerName != "" {
		podTemplateSpec.Spec.SchedulerName = cr.Spec.Config.SchedulerName
	}

	if cr.Spec.Config.Affinity != nil {
		podTemplateSpec.Spec.Affinity = cr.Spec.Config.Affinity
	}

	return nil
}

func UpdateSplunkPodTemplateWithConfig(podTemplateSpec *v1.PodTemplateSpec, cr *v1alpha1.SplunkEnterprise, instanceType enterprise.SplunkInstanceType) error {
	if cr.Spec.Config.DefaultsConfigMapName != "" {
		AddConfigMapVolumeToPodTemplate(podTemplateSpec, "splunk-defaults", cr.Spec.Config.DefaultsConfigMapName, enterprise.SPLUNK_DEFAULTS_MOUNT_LOCATION)
	}

	if cr.Spec.Config.SplunkLicense.VolumeSource != nil {
		AddLicenseVolumeToPodTemplate(podTemplateSpec, "splunk-license", cr.Spec.Config.SplunkLicense.VolumeSource, enterprise.LICENSE_MOUNT_LOCATION)
	}

	if cr.Spec.Config.EnableDFS && (instanceType == enterprise.SPLUNK_SEARCH_HEAD || instanceType == enterprise.SPLUNK_STANDALONE) {
		if err := AddDFCToPodTemplate(podTemplateSpec, cr); err != nil {
			return err
		}
	}

	return UpdatePodTemplateWithConfig(podTemplateSpec, cr)
}

func ParseResourceQuantity(str string, useIfEmpty string) (resource.Quantity, error) {
	var result resource.Quantity

	if (str == "") {
		if (useIfEmpty != "") {
			result = resource.MustParse(useIfEmpty)
		}
	} else {
		var err error
		result, err = resource.ParseQuantity(str)
		if err != nil {
			return result, fmt.Errorf("Invalid resource quantity \"%s\": %s", str, err)
		}
	}

	return result, nil
}


func GetSplunkVolumeMounts() ([]corev1.VolumeMount) {
	return []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "pvc-etc",
			MountPath: "/opt/splunk/etc",
		},
		corev1.VolumeMount{
			Name:      "pvc-var",
			MountPath: "/opt/splunk/var",
		},
	}
}


func GetSplunkVolumeClaims(cr *v1alpha1.SplunkEnterprise, instanceType enterprise.SplunkInstanceType, labels map[string]string) ([]corev1.PersistentVolumeClaim, error) {
	var err error
	var etcStorage, varStorage resource.Quantity

	etcStorage, err = ParseResourceQuantity(cr.Spec.Config.SplunkEtcStorage, "1Gi")
	if err != nil {
		return []corev1.PersistentVolumeClaim{}, fmt.Errorf("%s: %s", "splunkEtcStorage", err)
	}

	if (instanceType == enterprise.SPLUNK_INDEXER) {
		varStorage, err = ParseResourceQuantity(cr.Spec.Config.SplunkIndexerStorage, "200Gi")
		if err != nil {
			return []corev1.PersistentVolumeClaim{}, fmt.Errorf("%s: %s", "splunkIndexerStorage", err)
		}
	} else {
		varStorage, err = ParseResourceQuantity(cr.Spec.Config.SplunkVarStorage, "50Gi")
		if err != nil {
			return []corev1.PersistentVolumeClaim{}, fmt.Errorf("%s: %s", "splunkVarStorage", err)
		}
	}

	volumeClaims := []corev1.PersistentVolumeClaim{
		corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "etc",
				Namespace: cr.Namespace,
				Labels: labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: etcStorage,
					},
				},
			},
		},
		corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "var",
				Namespace: cr.Namespace,
				Labels: labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: varStorage,
					},
				},
			},
		},
	}

	if (cr.Spec.Config.StorageClassName != "") {
		for idx, _ := range volumeClaims {
			volumeClaims[idx].Spec.StorageClassName = &cr.Spec.Config.StorageClassName
		}
	}

	return volumeClaims, nil
}


func GetSplunkRequirements(cr *v1alpha1.SplunkEnterprise) (corev1.ResourceRequirements, error) {
	cpuRequest, err := ParseResourceQuantity(cr.Spec.Config.SplunkCpuRequest, "0.1")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkCpuRequest", err)
	}

	memoryRequest, err := ParseResourceQuantity(cr.Spec.Config.SplunkMemoryRequest, "512Mi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkMemoryRequest", err)
	}

	cpuLimit, err := ParseResourceQuantity(cr.Spec.Config.SplunkCpuLimit, "4")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkCpuLimit", err)
	}

	memoryLimit, err := ParseResourceQuantity(cr.Spec.Config.SplunkMemoryLimit, "8Gi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkMemoryLimit", err)
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuRequest,
			corev1.ResourceMemory: memoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		} }, nil
}


func GetSparkRequirements(cr *v1alpha1.SplunkEnterprise) (corev1.ResourceRequirements, error) {
	cpuRequest, err := ParseResourceQuantity(cr.Spec.Config.SparkCpuRequest, "0.1")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkCpuRequest", err)
	}

	memoryRequest, err := ParseResourceQuantity(cr.Spec.Config.SparkMemoryRequest, "512Mi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkMemoryRequest", err)
	}

	cpuLimit, err := ParseResourceQuantity(cr.Spec.Config.SparkCpuLimit, "4")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkCpuLimit", err)
	}

	memoryLimit, err := ParseResourceQuantity(cr.Spec.Config.SparkMemoryLimit, "8Gi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkMemoryLimit", err)
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuRequest,
			corev1.ResourceMemory: memoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		} }, nil
}
