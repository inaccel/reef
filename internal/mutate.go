package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func IsContainer(container corev1.Container, image string, index int, command string) bool {
	return container.Name == fmt.Sprintf("%s-%d", regexp.MustCompile("[^0-9A-Za-z]+").ReplaceAllString(image, "-"), index)
}

func Container(image string, index int, command string) corev1.Container {
	return corev1.Container{
		Name:  fmt.Sprintf("%s-%d", regexp.MustCompile("[^0-9A-Za-z]+").ReplaceAllString(image, "-"), index),
		Image: image,
		Args: strings.FieldsFunc(command, func(c rune) bool {
			return c == ' '
		}),
	}
}

func IsEnvVar(envVar corev1.EnvVar, name string) bool {
	return envVar.Name == strings.ReplaceAll(strcase.ToScreamingSnake(name), "/", "_")
}

func EnvVar(name, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  strings.ReplaceAll(strcase.ToScreamingSnake(name), "/", "_"),
		Value: value,
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

type PodDefaulter struct{}

func (PodDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("pod defaulter did not understand object: %T", obj)
	}

	for image, commands := range pod.Annotations {
		if strings.HasPrefix(image, "inaccel/") {
			for index, command := range strings.FieldsFunc(commands, func(c rune) bool {
				return c == '\n'
			}) {
				containerExists := false
				for i := range pod.Spec.InitContainers {
					if IsContainer(pod.Spec.InitContainers[i], image, index, command) {
						containerExists = true
						pod.Spec.InitContainers[i] = Container(image, index, command)
					}
				}
				if !containerExists {
					pod.Spec.InitContainers = append(pod.Spec.InitContainers, Container(image, index, command))
				}
			}
		}
	}

	for name, value := range pod.Labels {
		if strings.HasPrefix(name, "inaccel/") {
			for i := range pod.Spec.Containers {
				envVarExists := false
				for j := range pod.Spec.Containers[i].Env {
					if IsEnvVar(pod.Spec.Containers[i].Env[j], name) {
						envVarExists = true
						pod.Spec.Containers[i].Env[j] = EnvVar(name, value)
					}
				}
				if !envVarExists {
					pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, EnvVar(name, value))
				}
			}

			for i := range pod.Spec.InitContainers {
				envVarExists := false
				for j := range pod.Spec.InitContainers[i].Env {
					if IsEnvVar(pod.Spec.InitContainers[i].Env[j], name) {
						envVarExists = true
						pod.Spec.InitContainers[i].Env[j] = EnvVar(name, value)
					}
				}
				if !envVarExists {
					pod.Spec.InitContainers[i].Env = append(pod.Spec.InitContainers[i].Env, EnvVar(name, value))
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

	return nil
}
