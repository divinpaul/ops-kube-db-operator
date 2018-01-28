package postgres

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// MetricsExporter provides an abstraction for deploying k8s MetricsExporter deployment
type MetricsExporter struct {
	clientset kubernetes.Interface
}

// NewMetricsExporter returns new NewMetricsExporter for managing k8s MetricsExporter deployment
func NewMetricsExporter(clientset kubernetes.Interface) *MetricsExporter {
	return &MetricsExporter{
		clientset: clientset,
	}
}

// Deploy MetricsExporter k8s deployment
func (e *MetricsExporter) Deploy(namespace, crdName string) error {
	serviceName := fmt.Sprintf("%s-metrics-exporter", crdName)
	labels := getLabels(crdName)

	if err := e.applyConfigMap(labels, namespace, serviceName); nil != err {
		return err
	}

	if err := e.applyService(labels, namespace, serviceName, metricsExporterPort); nil != err {
		return err
	}

	return e.applyDeployment(labels, namespace, serviceName, metricsExporterPort)
}

func (e *MetricsExporter) applyConfigMap(labels map[string]string, namespace, name string) error {
	obj, err := e.clientset.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if _, err = e.clientset.CoreV1().ConfigMaps(namespace).Update(obj); nil != err {
			glog.Errorf("Error updating service %s in namespace %s:[%s]", namespace, name, err)

			return err
		}
	}

	if errors.IsNotFound(err) {
		if _, err = e.clientset.CoreV1().ConfigMaps(namespace).Create(updateConfigMap(&apiv1.ConfigMap{}, labels, namespace, name)); nil != err {
			glog.Errorf("Error creating config map %s in namespace %s:[%s]", namespace, name, err)

			return err
		}
	}

	return err
}

func updateConfigMap(cm *apiv1.ConfigMap, labels map[string]string, namespace, name string) *apiv1.ConfigMap {
	updateCommonObjectMeta(cm.GetObjectMeta(), labels, namespace, name)
	cm.GetObjectMeta().SetAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	cm.Data = map[string]string{"queries.yaml": exporterQueries}

	return cm
}

func (e *MetricsExporter) applyService(labels map[string]string, namespace, name string, port int) error {
	obj, err := e.clientset.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if _, err = e.clientset.CoreV1().Services(namespace).Update(updateService(obj, labels, namespace, name, port)); nil != err {
			glog.Errorf("Error updating service %s in namespace %s:[%s]", namespace, name, err)

			return err
		}
	}

	if errors.IsNotFound(err) {
		if _, err = e.clientset.CoreV1().Services(namespace).Create(updateService(&apiv1.Service{}, labels, namespace, name, port)); nil != err {
			glog.Errorf("Error creating service %s in namespace %s:[%s]", namespace, name, err)

			return err
		}
	}

	return err
}

func updateService(svc *apiv1.Service, labels map[string]string, namespace, name string, port int) *apiv1.Service {
	updateCommonObjectMeta(svc.GetObjectMeta(), labels, namespace, name)
	svc.GetObjectMeta().SetAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	svc.Spec = apiv1.ServiceSpec{
		Ports:    []apiv1.ServicePort{{Port: int32(port), TargetPort: intstr.FromInt(port)}},
		Selector: labels,
	}

	return svc
}

func (e *MetricsExporter) applyDeployment(labels map[string]string, namespace, name string, port int) error {
	obj, err := e.clientset.ExtensionsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		deployment := updateDeployment(obj, labels, namespace, name, port)
		if _, err = e.clientset.ExtensionsV1beta1().Deployments(namespace).Update(deployment); nil != err {
			glog.Errorf("Error updating deployment %s in namespace %s:[%s]", namespace, name, err)

			return err
		}
	}

	if errors.IsNotFound(err) {
		if _, err = e.clientset.ExtensionsV1beta1().Deployments(namespace).Create(updateDeployment(&extsv1beta1.Deployment{}, labels, namespace, name, port)); nil != err {
			glog.Errorf("Error creating deployment %s in namespace %s:[%s]", namespace, name, err)

			return err
		}
	}

	return err
}

func updateDeployment(deployment *extsv1beta1.Deployment, labels map[string]string, namespace, name string, port int) *extsv1beta1.Deployment {
	probe := &apiv1.Probe{
		Handler: apiv1.Handler{
			HTTPGet: &apiv1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(port),
			},
		},
		InitialDelaySeconds: 60,
		TimeoutSeconds:      3,
	}
	deploymentSpec := extsv1beta1.DeploymentSpec{
		Replicas:             int32Ptr(1),
		RevisionHistoryLimit: int32Ptr(2),
		Template: apiv1.PodTemplateSpec{
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{{
					Name:            "metrics",
					Image:           "wrouesnel/postgres_exporter:v0.4.1",
					ImagePullPolicy: "Always",
					Args:            []string{"--extend.query-path=/etc/config/queries.yaml"},
					Env: []apiv1.EnvVar{{
						Name: "DATA_SOURCE_NAME",
						ValueFrom: &apiv1.EnvVarSource{
							SecretKeyRef: &apiv1.SecretKeySelector{
								LocalObjectReference: apiv1.LocalObjectReference{Name: name},
								Key:                  "DATABASE_URL",
							},
						},
					}},
					Ports: []apiv1.ContainerPort{{
						Name:          "metrics",
						ContainerPort: int32(port),
					}},
					LivenessProbe:  probe,
					ReadinessProbe: probe,
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							"cpu":    quantity("100m"),
							"memory": quantity("256Mi"),
						},
					},
					VolumeMounts: []apiv1.VolumeMount{{
						Name:      "config-volume",
						MountPath: "/etc/config",
					}},
				}},
				Volumes: []apiv1.Volume{{
					Name: "config-volume",
					VolumeSource: apiv1.VolumeSource{
						ConfigMap: &apiv1.ConfigMapVolumeSource{
							LocalObjectReference: apiv1.LocalObjectReference{name},
						},
					},
				}},
			},
		},
	}
	updateCommonObjectMeta(deployment.GetObjectMeta(), labels, namespace, name)
	updateCommonObjectMeta(deploymentSpec.Template.GetObjectMeta(), labels, namespace, name)
	deployment.Spec = deploymentSpec

	return deployment
}

func updateCommonObjectMeta(objectMeta metav1.Object, labels map[string]string, namespace, name string) {
	objectMeta.SetName(name)
	objectMeta.SetNamespace(namespace)
	objectMeta.SetLabels(labels)
}

func getLabels(crdName string) map[string]string {
	labels := map[string]string{
		"app": "metrics-exporter",
		"crd": crdName,
	}

	return labels
}

func quantity(qty string) resource.Quantity {
	quantity, _ := resource.ParseQuantity(qty)

	return quantity
}

func int32Ptr(x int32) *int32 {
	return &x
}

const metricsExporterPort = 9187
const exporterQueries = `|
    pg_database:
      metrics:
      - datname:
          description: Name of the database
          usage: LABEL
      - size:
          description: Disk space used by the database
          usage: GAUGE
      query: ' SELECT pg_database.datname, pg_database_size(pg_database.datname) as size
        FROM pg_database'
    pg_postmaster:
      metrics:
      - start_time_seconds:
          description: Time at which postmaster started
          usage: GAUGE
      query: SELECT pg_postmaster_start_time as start_time_seconds from pg_postmaster_start_time()
    pg_replication:
      metrics:
      - lag:
          description: Replication lag behind master in seconds
          usage: GAUGE
      query: SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))::INT
        as lag
    pg_stat_user_tables:
      metrics:
      - schemaname:
          description: Name of the schema that this table is in
          usage: LABEL
      - relname:
          description: Name of this table
          usage: LABEL
      - seq_scan:
          description: Number of sequential scans initiated on this table
          usage: COUNTER
      - seq_tup_read:
          description: Number of live rows fetched by sequential scans
          usage: COUNTER
      - idx_scan:
          description: Number of index scans initiated on this table
          usage: COUNTER
      - idx_tup_fetch:
          description: Number of live rows fetched by index scans
          usage: COUNTER
      - n_tup_ins:
          description: Number of rows inserted
          usage: COUNTER
      - n_tup_upd:
          description: Number of rows updated
          usage: COUNTER
      - n_tup_del:
          description: Number of rows deleted
          usage: COUNTER
      - n_tup_hot_upd:
          description: Number of rows HOT updated (i.e., with no separate index update
            required)
          usage: COUNTER
      - n_live_tup:
          description: Estimated number of live rows
          usage: GAUGE
      - n_dead_tup:
          description: Estimated number of dead rows
          usage: GAUGE
      - n_mod_since_analyze:
          description: Estimated number of rows changed since last analyze
          usage: GAUGE
      - last_vacuum:
          description: Last time at which this table was manually vacuumed (not counting
            VACUUM FULL)
          usage: GAUGE
      - last_autovacuum:
          description: Last time at which this table was vacuumed by the autovacuum daemon
          usage: GAUGE
      - last_analyze:
          description: Last time at which this table was manually analyzed
          usage: GAUGE
      - last_autoanalyze:
          description: Last time at which this table was analyzed by the autovacuum daemon
          usage: GAUGE
      - vacuum_count:
          description: Number of times this table has been manually vacuumed (not counting
            VACUUM FULL)
          usage: COUNTER
      - autovacuum_count:
          description: Number of times this table has been vacuumed by the autovacuum
            daemon
          usage: COUNTER
      - analyze_count:
          description: Number of times this table has been manually analyzed
          usage: COUNTER
      - autoanalyze_count:
          description: Number of times this table has been analyzed by the autovacuum
            daemon
          usage: COUNTER
      query: SELECT schemaname, relname, seq_scan, seq_tup_read, idx_scan, idx_tup_fetch,
        n_tup_ins, n_tup_upd, n_tup_del, n_tup_hot_upd, n_live_tup, n_dead_tup, n_mod_since_analyze,
        last_vacuum, last_autovacuum, last_analyze, last_autoanalyze, vacuum_count, autovacuum_count,
        analyze_count, autoanalyze_count FROM pg_stat_user_tables
`
