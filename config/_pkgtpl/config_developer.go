// +build ignore

package developer

import (
	"github.com/corestoreio/csfw/config/element"
	"github.com/corestoreio/csfw/store/scope"
)

// ConfigStructure global configuration structure for this package.
// Used in frontend and backend. See init() for details.
var ConfigStructure element.SectionSlice

func init() {
	ConfigStructure = element.MustNewConfiguration(
		element.Section{
			ID: "dev",
			Groups: element.NewGroupSlice(
				element.Group{
					ID:        "front_end_development_workflow",
					Label:     `Frontend Development Workflow`,
					SortOrder: 8,
					Scopes:    scope.PermStore,
					Fields: element.NewFieldSlice(
						element.Field{
							// Path: dev/front_end_development_workflow/type
							ID:        "type",
							Label:     `Workflow type`,
							Comment:   text.Long(`Not available in production mode`),
							Type:      element.TypeSelect,
							SortOrder: 1,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermDefault,
							Default:   `server_side_compilation`,
							// SourceModel: Magento\Developer\Model\Config\Source\WorkflowType
						},
					),
				},

				element.Group{
					ID:        "restrict",
					Label:     `Developer Client Restrictions`,
					SortOrder: 10,
					Scopes:    scope.PermStore,
					Fields: element.NewFieldSlice(
						element.Field{
							// Path: dev/restrict/allow_ips
							ID:        "allow_ips",
							Label:     `Allowed IPs (comma separated)`,
							Comment:   text.Long(`Leave empty for access from any location.`),
							Type:      element.TypeText,
							SortOrder: 20,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermStore,
							// BackendModel: Magento\Developer\Model\Config\Backend\AllowedIps
						},
					),
				},
			),
		},

		// Hidden Configuration, may be visible somewhere else ...
		element.Section{
			ID: "dev",
			Groups: element.NewGroupSlice(
				element.Group{
					ID: "restrict",
					Fields: element.NewFieldSlice(
						element.Field{
							// Path: dev/restrict/allow_ips
							ID:      `allow_ips`,
							Type:    element.TypeHidden,
							Visible: element.VisibleNo,
						},
					),
				},
			),
		},
	)
	Backend = NewBackend(ConfigStructure)
}
