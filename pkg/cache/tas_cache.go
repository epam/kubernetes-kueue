/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"maps"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
)

type TASCache struct {
	sync.RWMutex
	client     client.Client
	topologies map[kueue.TopologyReference][]string
	flavors    map[kueue.ResourceFlavorReference]*TASFlavorCache
}

func NewTASCache(client client.Client) TASCache {
	return TASCache{
		client:     client,
		topologies: make(map[kueue.TopologyReference][]string),
		flavors:    make(map[kueue.ResourceFlavorReference]*TASFlavorCache),
	}
}

func (t *TASCache) GetTopology(name kueue.TopologyReference) ([]string, bool) {
	t.RLock()
	defer t.RUnlock()
	levels, ok := t.topologies[name]
	return levels, ok
}

// CloneTopologies returns a shallow copy of the map
func (t *TASCache) CloneTopologies() map[kueue.TopologyReference][]string {
	t.RLock()
	defer t.RUnlock()
	return maps.Clone(t.topologies)
}

func (t *TASCache) SetTopology(name kueue.TopologyReference, levels []string) {
	t.Lock()
	defer t.Unlock()
	t.topologies[name] = levels
}

func (t *TASCache) DeleteTopology(name kueue.TopologyReference) {
	t.Lock()
	defer t.Unlock()
	delete(t.topologies, name)
}

func (t *TASCache) GetFlavor(name kueue.ResourceFlavorReference) *TASFlavorCache {
	t.RLock()
	defer t.RUnlock()
	return t.flavors[name]
}

// CloneFlavors returns a shallow copy of the map
func (t *TASCache) CloneFlavors() map[kueue.ResourceFlavorReference]*TASFlavorCache {
	t.RLock()
	defer t.RUnlock()
	return maps.Clone(t.flavors)
}

func (t *TASCache) SetFlavor(name kueue.ResourceFlavorReference, info *TASFlavorCache) {
	t.Lock()
	defer t.Unlock()
	t.flavors[name] = info
}

func (t *TASCache) DeleteFlavor(name kueue.ResourceFlavorReference) {
	t.Lock()
	defer t.Unlock()
	delete(t.flavors, name)
}
