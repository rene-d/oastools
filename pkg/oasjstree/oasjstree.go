package oasjstree

import (
	"log"

	"../../pkg/oasmodel"
)

const iconPath = "./dist/oas/"
const iconObject = iconPath + "object.png"
const iconArray = iconPath + "array.png"
const iconBool = iconPath + "bool.png"
const iconInt = iconPath + "int.png"
const iconString = iconPath + "string.png"

/*jstreeNode
 Specs:
 // Expected format of the node (there are no required fields)
{
  id          : "string" // will be autogenerated if omitted
  text        : "string" // node text
  icon        : "string" // string for custom
  state       : {
    opened    : boolean  // is the node open
    disabled  : boolean  // is the node disabled
    selected  : boolean  // is the node selected
  },
  children    : []  // array of strings or objects
  li_attr     : {}  // attributes for the generated LI node
  a_attr      : {}  // attributes for the generated A node
}
*/

//JstNodeState : options for jstreeNode
type JstNodeState struct {
	Opened   bool `json:"opened,omitempty"`   // is the node open
	Disabled bool `json:"disable,omitempty"`  // is the node disabled
	Selected bool `json:"selected,omitempty"` // is the node selected
}

//JstNode : jstree Node
type JstNode struct {
	ID       string        `json:"id,omitempty"`
	Text     string        `json:"text,omitempty"`     // node text
	Icon     string        `json:"icon,omitempty"`     // string for custom icon
	State    *JstNodeState `json:"state,omitempty"`    // options
	Children []*JstNode    `json:"children,omitempty"` // array of strings or objects
	// unused Li_attr     : {}  // attributes for the generated LI node
	// unused A_attr      : {}  // attributes for the generated A node
}

// Generator Structures

// OASType : Generic OAS objects
type OASType interface {
	GetNode(unfold bool) *JstNode
	UpdateTypeRefs(refs map[string]OASType)
}

//OASTypeRef behavior when a OAS Object is a reference
type OASTypeRef struct {
	name      string
	refString string
	TypeRef   OASType
}

//OASArray : Array of items
type OASArray struct {
	name  string
	Items OASType
}

//objectMember : object's member
type objectMember struct {
	Type OASType
	Name string
}

//OASObject : object structure
type OASObject struct {
	name    string
	Members map[string]objectMember
}

//OASBasicTypeMember : just a string representing type name
type OASBasicTypeMember struct {
	name     string
	typeName string
}

//GetNode : OASTypeRef OASType interface realization
func (tr *OASTypeRef) GetNode(unfold bool) *JstNode {
	var node JstNode
	node = *tr.TypeRef.GetNode(unfold)
	node.Text = tr.name

	return &node
}

//UpdateTypeRefs : OASTypeRef OASType interface realization
func (tr *OASTypeRef) UpdateTypeRefs(refs map[string]OASType) {
	tr.TypeRef = refs[tr.refString]
}

//GetNode : OASArray OASType interface realization
func (a *OASArray) GetNode(unfold bool) *JstNode {
	node := JstNode{"", a.name, iconArray, &JstNodeState{unfold, false, false}, nil}
	node.Children = append(node.Children, a.Items.GetNode(unfold))
	return &node
}

//UpdateTypeRefs : OASArray OASType interface realization
func (a *OASArray) UpdateTypeRefs(refs map[string]OASType) {
	a.Items.UpdateTypeRefs(refs)
}

//GetNode : OASArray OASType interface realization
func (o *OASObject) GetNode(unfold bool) *JstNode {
	node := JstNode{"", o.name, iconObject, &JstNodeState{unfold, false, false}, nil}
	log.Printf("new object node : %s #members:%d\n", o.name, len(o.Members))
	for m := range o.Members {
		log.Printf("add Member:%s to object:%s\n", o.Members[m].Name, o.name)
		node.Children = append(node.Children, o.Members[m].Type.GetNode(unfold))
	}
	return &node
}

//UpdateTypeRefs : OASArray OASType interface realization
func (o *OASObject) UpdateTypeRefs(refs map[string]OASType) {
	for c := range o.Members {
		o.Members[c].Type.UpdateTypeRefs(refs)
	}
}

//GetNode : OASBasicTypeMember OASType interface realization
func (btm *OASBasicTypeMember) GetNode(unfold bool) *JstNode {
	node := JstNode{"", btm.name, "", &JstNodeState{unfold, false, false}, nil}
	switch btm.typeName {
	case "boolean":
		node.Icon = iconBool
	case "integer":
		node.Icon = iconInt
	case "string":
		node.Icon = iconString
	}
	return &node
}

//UpdateTypeRefs : OASBasicTypeMember OASType interface realization
func (btm *OASBasicTypeMember) UpdateTypeRefs(refs map[string]OASType) {
}

func memberToHTML(m string, desc string) string {
	return "<table><tr><td width=\"200px\">" + m + "</td><td>: " + desc + "</td></tr></table>"
}

//SchemaToJst : convert OAS DataModel to Jst Ready model generator
func SchemaToJst(name string, schema oasmodel.SchemaOrRef) OASType {
	if schema.Ref != nil {
		Ret := OASTypeRef{name, schema.Ref.Ref, nil}
		return &Ret
	}

	/* object */
	if schema.Val.Type == "object" {
		var o OASObject
		o.name = name
		o.Members = make(map[string]objectMember)

		for m := range schema.Val.Properties {
			mtype := SchemaToJst(m, *schema.Val.Properties[m])
			var desc string
			if schema.Val.Properties[m].Val != nil {
				desc = schema.Val.Properties[m].Val.Description
			}
			o.Members[m] = objectMember{mtype, memberToHTML(m, desc)}
		}
		return &o
	}

	/* array */
	if schema.Val.Type == "array" {
		var a OASArray
		a.name = name
		a.Items = SchemaToJst(".", *schema.Val.Items)

		return &a
	}
	/* Basic Type */
	Ret := OASBasicTypeMember{name, schema.Val.Type}
	return &Ret
}

//GetJstree : return jstree root data info
func GetJstree(file string, object string, unfold bool) *JstNode {
	// First load the file against OAS Specs
	oa := oasmodel.OpenAPI{}
	oa.Load(file)
	return getJstree(&oa, object, unfold)
}

//GetJstreeFromData : return jstree root data info
func GetJstreeFromData(buffer []byte, object string, unfold bool) *JstNode {
	oa := oasmodel.OpenAPI{}
	oa.UnMarshal(buffer)
	return getJstree(&oa, object, unfold)
}

func getJstree(oa *oasmodel.OpenAPI, object string, unfold bool) *JstNode {
	// Generate Jst Internal Generator Representation
	components := make(map[string]OASType)
	Refs := make(map[string]OASType)
	for k := range oa.Components.Schemas {
		elem := oa.Components.Schemas[k]
		component := SchemaToJst(k, *elem)
		components[k] = component
		Refs["#/components/schemas/"+k] = component
	}
	// Update types references
	for i := range components {
		components[i].UpdateTypeRefs(Refs)
	}

	if components[object] == nil {
		return nil
	}
	return components[object].GetNode(unfold)
}
