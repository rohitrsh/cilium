// Copyright 2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"strings"

	"github.com/cilium/cilium/pkg/ipam/types"
)

const (
	// ProviderPrefix is the prefix used to indicate that a k8s ProviderID
	// represents an Azure resource
	ProviderPrefix = "azure://"

	// InterfaceAddressLimit is the maximum number of addresses on an interface
	//
	//
	// For more information:
	// https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/azure-subscription-service-limits?toc=%2fazure%2fvirtual-network%2ftoc.json#networking-limits
	InterfaceAddressLimit = 256

	// StateSucceeded is the address state for a successfully provisioned address
	StateSucceeded = "succeeded"
)

// AzureStatus is the status of Azure addressing of the node
//
// This struct is embedded into v2.CiliumNode
//
// +k8s:deepcopy-gen=true
type AzureStatus struct {
	// Interfaces is the list of interfaces on the node
	//
	// +optional
	Interfaces []AzureInterface `json:"interfaces,omitempty"`
}

// AzureAddress is an IP address assigned to an AzureInterface
type AzureAddress struct {
	// IP is the ip address of the address
	IP string `json:"ip,omitempty"`

	// Subnet is the subnet the address belongs to
	Subnet string `json:"subnet,omitempty"`

	// State is the provisioning state of the address
	State string `json:"state,omitempty"`
}

// AzureInterface represents an Azure Interface
//
// +k8s:deepcopy-gen=true
type AzureInterface struct {
	// ID is the identifier
	//
	// +optional
	ID string `json:"id,omitempty"`

	// Name is the name of the interface
	//
	// +optional
	Name string `json:"name,omitempty"`

	// MAC is the mac address
	//
	// +optional
	MAC string `json:"mac,omitempty"`

	// State is the provisioning state
	//
	// +optional
	State string `json:"state,omitempty"`

	// Addresses is the list of all IPs associated with the interface,
	// including all secondary addresses
	//
	// +optional
	Addresses []AzureAddress `json:"addresses,omitempty"`

	// SecurityGroup is the security group associated with the interface
	SecurityGroup string `json:"security-group,omitempty"`

	// vmssName is the name of the virtual machine scale set. This field is
	// set by extractIDs()
	vmssName string

	// vmName is the name of the virtual machine
	vmName string

	// resourceGroup is the resource group the interface belongs to
	resourceGroup string
}

// InterfaceID returns the identifier of the interface
func (a *AzureInterface) InterfaceID() string {
	return a.ID
}

func (a *AzureInterface) extractIDs() {
	switch {
	// //subscriptions/xxx/resourceGroups/yyy/providers/Microsoft.Compute/virtualMachineScaleSets/ssss/virtualMachines/vvv/networkInterfaces/iii
	case strings.Contains(a.ID, "virtualMachineScaleSets"):
		segs := strings.Split(a.ID, "/")
		if len(segs) >= 5 {
			a.resourceGroup = segs[4]
		}
		if len(segs) >= 9 {
			a.vmssName = segs[8]
		}
		if len(segs) >= 11 {
			a.vmName = segs[10]
		}
	}
}

// ResourceGroup returns the resource group the interface belongs to
func (a *AzureInterface) ResourceGroup() string {
	if a.resourceGroup == "" {
		a.extractIDs()
	}
	return a.resourceGroup
}

// VMScaleSetName returns the VM scale set name the interface belongs to
func (a *AzureInterface) VMScaleSetName() string {
	if a.vmssName == "" {
		a.extractIDs()
	}
	return a.vmssName
}

// VMName returns the VM name the interface belongs to
func (a *AzureInterface) VMName() string {
	if a.vmName == "" {
		a.extractIDs()
	}
	return a.vmName
}

// ForeachAddress iterates over all addresses and calls fn
func (a *AzureInterface) ForeachAddress(id string, fn types.AddressIterator) error {
	for _, address := range a.Addresses {
		if err := fn(id, a.ID, address.IP, address.Subnet, address); err != nil {
			return err
		}
	}

	return nil
}
