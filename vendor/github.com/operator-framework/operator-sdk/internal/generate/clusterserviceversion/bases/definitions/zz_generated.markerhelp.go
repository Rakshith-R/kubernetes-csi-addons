//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2020 The Operator-SDK Authors
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

// Code generated by helpgen. DO NOT EDIT.

package definitions

import (
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func (Description) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "Description",
		DetailedHelp: markers.DetailedHelp{
			Summary: "is a type used to receive type-level CRD description markers.",
			Details: "",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"Resources": markers.DetailedHelp{
				Summary: "is a list of string lists, each of which defines a CRD description resource. The marker format is: { { \"kind\" , \"version\" ( , \"name\")? } , ... }",
				Details: "",
			},
			"DisplayName": markers.DetailedHelp{
				Summary: "is the displayName of a CRD description.",
				Details: "",
			},
			"Order": markers.DetailedHelp{
				Summary: "determines which position in the list this description will take. Markers with Order omitted have the highest Order. If more than one marker has the same Order, the corresponding descriptions will be sorted alphabetically and placed above others with higher Orders.",
				Details: "",
			},
		},
	}
}

func (Descriptor) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "Descriptor",
		DetailedHelp: markers.DetailedHelp{
			Summary: "is a type used to receive field-level spec and status descriptor markers. Format of marker:",
			Details: "",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"Type": markers.DetailedHelp{
				Summary: "is one of: \"spec\", \"status\".",
				Details: "",
			},
			"DisplayName": markers.DetailedHelp{
				Summary: "is the displayName of a spec or status description.",
				Details: "",
			},
			"XDescriptors": markers.DetailedHelp{
				Summary: "is a list of UI path strings. The marker format is: \"ui:element:foo,ui:element:bar\"",
				Details: "",
			},
			"Order": markers.DetailedHelp{
				Summary: "determines which position in the list this descriptor will take. Markers with Order omitted have the highest Order. If more than one marker has the same Order, the corresponding descriptors will be sorted alphabetically and placed above others with higher Orders.",
				Details: "",
			},
		},
	}
}
