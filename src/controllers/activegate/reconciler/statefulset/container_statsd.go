package statefulset

import (
	"fmt"

	"github.com/Dynatrace/dynatrace-operator/src/controllers/activegate/internal/consts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const statsDProbesPortName = "statsd-probes"
const statsDProbesPort = 14999

var _ ContainerBuilder = (*StatsD)(nil)

type StatsD struct {
	GenericContainer
}

func NewStatsD(stsProperties *statefulSetProperties) *StatsD {
	return &StatsD{
		GenericContainer: *NewGenericContainer(stsProperties),
	}
}

func (statsd *StatsD) BuildContainer() corev1.Container {
	return corev1.Container{
		Name:            consts.StatsDContainerName,
		Image:           statsd.StsProperties.DynaKube.ActiveGateImage(),
		ImagePullPolicy: corev1.PullAlways,
		Env:             statsd.buildEnvs(),
		VolumeMounts:    statsd.buildVolumeMounts(),
		Command:         statsd.buildCommand(),
		Ports:           statsd.buildPorts(),
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/readyz",
					Port: intstr.IntOrString{IntVal: statsDProbesPort},
				},
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       15,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/livez",
					Port: intstr.IntOrString{IntVal: statsDProbesPort},
				},
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       15,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
		},
	}
}

func (statsd *StatsD) BuildVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "ds-metadata",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func (statsd *StatsD) buildCommand() []string {
	if statsd.StsProperties.DynaKube.FeatureUseActiveGateImageForStatsD() {
		return []string{
			"/bin/bash", "/dt/statsd/entrypoint.sh",
		}
	}
	return nil
}

func (statsd *StatsD) buildPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{Name: consts.StatsDIngestTargetPort, ContainerPort: consts.StatsDIngestPort},
		{Name: statsDProbesPortName, ContainerPort: statsDProbesPort},
	}
}

func (statsd *StatsD) buildVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{Name: "eec-ds-shared", MountPath: "/mnt/dsexecargs"},
		{Name: "dsauthtokendir", MountPath: "/var/lib/dynatrace/remotepluginmodule/agent/runtime/datasources"},
		{Name: "ds-metadata", MountPath: "/mnt/dsmetadata"},
	}
}

func (statsd *StatsD) buildEnvs() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "StatsDExecArgsPath", Value: "/mnt/dsexecargs/statsd.process.json"},
		{Name: "ProbeServerPort", Value: fmt.Sprintf("%d", statsDProbesPort)},
		{Name: "StatsDMetadataDir", Value: "/mnt/dsmetadata"},
	}
}
