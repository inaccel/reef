package internal

import (
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func IsInitContainer(initContainer corev1.Container, image string, index int, command string) bool {
	return initContainer.Name == fmt.Sprintf("%s-%d", regexp.MustCompile("[^0-9A-Za-z]+").ReplaceAllString(image, "-"), index)
}

func InitContainer(image string, index int, command string) corev1.Container {
	return corev1.Container{
		Name:  fmt.Sprintf("%s-%d", regexp.MustCompile("[^0-9A-Za-z]+").ReplaceAllString(image, "-"), index),
		Image: image,
		Args: strings.FieldsFunc(command, func(c rune) bool {
			return c == ' '
		}),
	}
}

func IsVolume(volume corev1.Volume) bool {
	return volume.Name == "inaccel"
}

func Volume() corev1.Volume {
	return corev1.Volume{
		Name: "inaccel",
		VolumeSource: corev1.VolumeSource{
			CSI: &corev1.CSIVolumeSource{
				Driver: "inaccel",
			},
		},
	}
}

func IsVolumeMount(volumeMount corev1.VolumeMount) bool {
	return volumeMount.Name == "inaccel"
}

func VolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "inaccel",
		MountPath: "/var/lib/inaccel",
	}
}

func Mutate(me corev1.Pod) (corev1.Pod, error) {
	var pod corev1.Pod
	me.DeepCopyInto(&pod)

	for image, commands := range pod.Annotations {
		if strings.HasPrefix(image, "inaccel/") {
			for index, command := range strings.FieldsFunc(commands, func(c rune) bool {
				return c == '\n'
			}) {
				initContainerExists := false
				for i := range pod.Spec.InitContainers {
					if IsInitContainer(pod.Spec.InitContainers[i], image, index, command) {
						initContainerExists = true
						pod.Spec.InitContainers[i] = InitContainer(image, index, command)
					}
				}
				if !initContainerExists {
					pod.Spec.InitContainers = append(pod.Spec.InitContainers, InitContainer(image, index, command))
				}
			}
		}
	}

	for i := range pod.Spec.Containers {
		volumeMountExists := false
		for j := range pod.Spec.Containers[i].VolumeMounts {
			if IsVolumeMount(pod.Spec.Containers[i].VolumeMounts[j]) {
				volumeMountExists = true
				pod.Spec.Containers[i].VolumeMounts[j] = VolumeMount()
			}
		}
		if !volumeMountExists {
			pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, VolumeMount())
		}
	}

	for i := range pod.Spec.InitContainers {
		volumeMountExists := false
		for j := range pod.Spec.InitContainers[i].VolumeMounts {
			if IsVolumeMount(pod.Spec.InitContainers[i].VolumeMounts[j]) {
				volumeMountExists = true
				pod.Spec.InitContainers[i].VolumeMounts[j] = VolumeMount()
			}
		}
		if !volumeMountExists {
			pod.Spec.InitContainers[i].VolumeMounts = append(pod.Spec.InitContainers[i].VolumeMounts, VolumeMount())
		}
	}

	volumeExists := false
	for i := range pod.Spec.Volumes {
		if IsVolume(pod.Spec.Volumes[i]) {
			volumeExists = true
			pod.Spec.Volumes[i] = Volume()
		}
	}
	if !volumeExists {
		pod.Spec.Volumes = append(pod.Spec.Volumes, Volume())
	}

	return pod, nil
}
