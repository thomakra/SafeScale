/*
 * Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"fmt"
	"log"

	brokerclient "github.com/CS-SI/SafeScale/broker/client"

	clusterapi "github.com/CS-SI/SafeScale/deploy/cluster/api"
	"github.com/CS-SI/SafeScale/deploy/cluster/enums/Flavor"
	"github.com/CS-SI/SafeScale/deploy/cluster/flavors/boh"
	"github.com/CS-SI/SafeScale/deploy/cluster/flavors/dcos"
	"github.com/CS-SI/SafeScale/deploy/cluster/flavors/k8s"
	"github.com/CS-SI/SafeScale/deploy/cluster/flavors/ohpc"
	"github.com/CS-SI/SafeScale/deploy/cluster/metadata"
)

// Get returns the Cluster instance corresponding to the cluster named 'name'
func Get(name string) (clusterapi.Cluster, error) {
	m, err := metadata.NewCluster()
	if err != nil {
		return nil, err
	}
	found, err := m.Read(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get information about Cluster '%s': %s", name, err.Error())
	}
	if !found {
		return nil, nil
	}

	var instance clusterapi.Cluster
	common := m.Get()
	switch common.Flavor {
	case Flavor.DCOS:
		instance, err = dcos.Load(m)
		if err != nil {
			return nil, err
		}
	case Flavor.K8S:
		instance, err = k8s.Load(m)
		if err != nil {
			return nil, err
		}
	case Flavor.BOH:
		instance, err = boh.Load(m)
		if err != nil {
			return nil, err
		}
	case Flavor.OHPC:
		instance, err = ohpc.Load(m)
		if err != nil {
			return nil, err
		}
	default:
		found = false
	}
	if !found {
		return nil, nil
	}
	return instance, nil
}

// Create creates a cluster following the parameters of the request
func Create(req clusterapi.Request) (clusterapi.Cluster, error) {
	// Validates parameters
	if req.Name == "" {
		panic("req.Name is empty!")
	}
	if req.CIDR == "" {
		panic("req.CIDR is empty!")
	}

	var instance clusterapi.Cluster

	log.Printf("Creating infrastructure for cluster '%s'", req.Name)

	tenant, err := brokerclient.New().Tenant.Get(0)
	if err != nil {
		return nil, err
	}

	switch req.Flavor {
	case Flavor.DCOS:
		req.Tenant = tenant.Name
		instance, err = dcos.Create(req)
		if err != nil {
			return nil, err
		}
	case Flavor.BOH:
		req.Tenant = tenant.Name
		instance, err = boh.Create(req)
		if err != nil {
			return nil, err
		}
	case Flavor.OHPC:
		req.Tenant = tenant.Name
		instance, err = ohpc.Create(req)
		if err != nil {
			return nil, err
		}
	case Flavor.K8S:
		req.Tenant = tenant.Name
		instance, err = k8s.Create(req)
		if err != nil {
			return nil, err
		}
	//case Flavor.Swarm:
	default:
		return nil, fmt.Errorf("cluster Flavor '%s' not yet implemented", req.Flavor.String())
	}

	log.Printf("Cluster '%s' created and initialized successfully", req.Name)
	return instance, nil
}

// Delete deletes the infrastructure of the cluster named 'name'
func Delete(name string) error {
	instance, err := Get(name)
	if err != nil {
		return fmt.Errorf("failed to find a cluster named '%s': %s", name, err.Error())
	}
	if instance == nil {
		return fmt.Errorf("Cluster '%s' not found", name)
	}

	// Deletes all the infrastructure built for the cluster
	return instance.Delete()
}

// List lists the clusters already created
func List() ([]clusterapi.Cluster, error) {
	var clusterList []clusterapi.Cluster
	m, err := metadata.NewCluster()
	if err != nil {
		return clusterList, err
	}
	var instance clusterapi.Cluster
	err = m.Browse(func(cm *metadata.Cluster) error {
		cluster := cm.Get()
		switch cluster.Flavor {
		case Flavor.DCOS:
			instance, err = dcos.Load(cm)
			if err != nil {
				return err
			}
		case Flavor.BOH:
			instance, err = boh.Load(cm)
			if err != nil {
				return err
			}
		case Flavor.K8S:
			instance, err = k8s.Load(cm)
			if err != nil {
				return err
			}
		case Flavor.OHPC:
			fallthrough
		case Flavor.Swarm:
			return fmt.Errorf("cluster Flavor '%s' not yet implemented", cluster.Flavor.String())
		}

		clusterList = append(clusterList, instance)
		return nil
	})
	return clusterList, err
}
