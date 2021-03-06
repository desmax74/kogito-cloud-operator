// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
)

// WatchedObjects objects that the controller supposed to watch for
type WatchedObjects struct {
	// GroupVersion for the watched objects
	GroupVersion schema.GroupVersion
	// AddToScheme function to register the Scheme to the Kubernetes Client. This will enable the controller to query for those objects during the reconcile loop.
	AddToScheme func(scheme *runtime.Scheme) error
	// Objects list of required objects that should be watched by the controller
	Objects []runtime.Object
	// Owner of the object if different from the actual controller
	Owner runtime.Object
}

// ControllerWatcher helps to add required objects to the controller watch list given the required runtime objects
type ControllerWatcher interface {
	// Watch add the given objects to the controller watch list
	Watch(objects ...WatchedObjects) (err error)
	// AreAllObjectsWatched verifies if this instance already has registered every required object in the watch list
	AreAllObjectsWatched() bool
	// IsGroupWatched verifies if the given group has it's objects watched or not
	IsGroupWatched(group string) bool
}

// NewControllerWatcher creates a new ControllerWatcher to control the objects that needed to be watched
func NewControllerWatcher(client *client.Client, manager controllerruntime.Manager, controller controller.Controller, owner runtime.Object) ControllerWatcher {
	return &controllerWatcher{
		client:           client,
		manager:          manager,
		controller:       controller,
		owner:            owner,
		groupsNotWatched: map[string]bool{},
	}
}

type controllerWatcher struct {
	client           *client.Client
	manager          controllerruntime.Manager
	controller       controller.Controller
	owner            runtime.Object
	groupsNotWatched map[string]bool
}

func (c *controllerWatcher) AreAllObjectsWatched() bool {
	return len(c.groupsNotWatched) == 0
}

func (c *controllerWatcher) IsGroupWatched(group string) bool {
	if len(c.groupsNotWatched) == 0 {
		return true
	}
	_, exists := c.groupsNotWatched[group]
	return !exists
}

func (c *controllerWatcher) Watch(watchedObjects ...WatchedObjects) (err error) {
	serverGroups, err := c.client.Discovery.ServerGroups()
	if err != nil {
		return
	}

	var addToScheme runtime.SchemeBuilder
	var desiredObjects []WatchedObjects

	for _, object := range watchedObjects {
		// core resources
		if object.AddToScheme == nil {
			desiredObjects = append(desiredObjects, object)
		} else {
			found := false
			for _, serverGroup := range serverGroups.Groups {
				if strings.Contains(serverGroup.Name, object.GroupVersion.Group) {
					addToScheme = append(addToScheme, object.AddToScheme)
					desiredObjects = append(desiredObjects, object)
					found = true
					delete(c.groupsNotWatched, object.GroupVersion.Group)
					break
				}
			}
			if !found {
				c.groupsNotWatched[object.GroupVersion.Group] = true
				log.Warnf("Impossible to register GroupVersion %s. CRD not installed in the cluster, controller might not behave as expected", object.GroupVersion)
			}
		}
	}

	if len(addToScheme) > 0 {
		log.Debug("Registering additional controller schemes")
		if err = addToScheme.AddToScheme(c.manager.GetScheme()); err != nil {
			return
		}
	}

	ownerHandler := &handler.EnqueueRequestForOwner{IsController: true, OwnerType: c.owner}
	for _, desiredObject := range desiredObjects {
		for _, runtimeObj := range desiredObject.Objects {
			if desiredObject.Owner == nil {
				if err = c.controller.Watch(&source.Kind{Type: runtimeObj}, ownerHandler); err != nil {
					return
				}
			} else {
				if err = c.controller.Watch(
					&source.Kind{Type: runtimeObj},
					&handler.EnqueueRequestForOwner{IsController: true, OwnerType: desiredObject.Owner}); err != nil {
					return
				}
			}
		}
	}

	return
}
