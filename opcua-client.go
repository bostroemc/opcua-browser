package main

import (
	"context"
	"log"
	"time"

	"github.com/bostroemc/tui/opcua-browser/types"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func opcuaClient(config types.Config, browse chan types.OpcUaBrowserData, read chan types.OpcUaReadData, write chan types.DataPoint) {
	ctx := context.Background()

	endpoints, err := opcua.GetEndpoints(ctx, config.Server.Endpoint)
	if err != nil {
		log.Println(err)
	}
	ep, err := opcua.SelectEndpoint(endpoints, config.Server.Policy, ua.MessageSecurityModeFromString(config.Server.Mode))
	if err != nil {
		log.Println(err)
		return
	}
	ep.EndpointURL = config.Server.Endpoint

	opts := []opcua.Option{
		opcua.SecurityPolicy(config.Server.Policy),
		opcua.SecurityModeString(config.Server.Mode),
		opcua.CertificateFile(config.Authorization.Certificate),
		opcua.PrivateKeyFile(config.Authorization.Key),
		opcua.AuthUsername(config.Authorization.Username, config.Authorization.Password),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeUserName),
		opcua.SessionTimeout(30 * time.Second),
	}

	c, err := opcua.NewClient(ep.EndpointURL, opts...)
	if err != nil {
		log.Println(err)
	}
	if err := c.Connect(ctx); err != nil {
		log.Println(err)
	}

	// defer c.Close(ctx)  //TODO figure out how to Close properly...

	go func() {
		for {
			a := <-browse

			if isActive(ctx, c) == false {
				time.Sleep(1 * time.Second)
			}

			refs, err := c.Node(a.Node).ReferencedNodes(ctx, 0, ua.BrowseDirectionForward, ua.NodeClassAll, true)
			if err != nil { //TODO handle error case; check whether error occurs in case there are no referenced nodes (i.e. there are no children)
			}
			var parent types.Node
			attrs, _ := c.Node(a.Node).Attributes(ctx, ua.AttributeIDNodeID, ua.AttributeIDBrowseName, ua.AttributeIDDescription, ua.AttributeIDAccessLevel, ua.AttributeIDDataType)
			parent = types.Node{NodeID: attrs[0].Value.NodeID(), BrowseName: attrs[1].Value.String(), Description: attrs[2].Value.String(), DataType: attrs[4].Value.String()}

			var children []types.Node
			for _, s := range refs {
				attrs, _ := s.Attributes(ctx, ua.AttributeIDNodeID, ua.AttributeIDBrowseName, ua.AttributeIDDescription, ua.AttributeIDAccessLevel, ua.AttributeIDDataType)
				children = append(children, types.Node{NodeID: attrs[0].Value.NodeID(), BrowseName: attrs[1].Value.String(), Description: attrs[2].Value.String(), DataType: attrs[4].Value.String()})
			}

			browse <- types.OpcUaBrowserData{Parent: parent, Children: children}
		}
	}()

	//write
	go func() {
		for {
			x := <-write

			_temp := x.Node
			id_1, _ := ua.ParseNodeID(_temp)
			v, _ := ua.NewVariant(x.Pending)
			req := &ua.WriteRequest{
				NodesToWrite: []*ua.WriteValue{
					{
						NodeID:      id_1,
						AttributeID: ua.AttributeIDValue,
						Value: &ua.DataValue{
							EncodingMask: ua.DataValueValue,
							Value:        v,
						},
					},
				},
			}
			if c.State() != opcua.Connected {
				continue
			}
			_, err := c.Write(ctx, req)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	go func() {
		var resp *ua.ReadResponse
		for {
			a := <-read
			// time.Sleep(5 * time.Second)

			var Nodes []*ua.ReadValueID
			for _, d := range a.Data {
				_id, _ := ua.ParseNodeID(d.Node)
				Nodes = append(Nodes, &ua.ReadValueID{NodeID: _id})
			}

			if len(Nodes) > 0 {
				req := ua.ReadRequest{NodesToRead: Nodes}

				resp, err = c.Read(ctx, &req)
				if err != nil {
					read <- types.OpcUaReadData{}
					continue
				}
				for i, r := range resp.Results {
					a.Data[i].Value = r.Value.Value()
				}
			}
			read <- a
		}
	}()
}

func isActive(ctx context.Context, client *opcua.Client) bool {
	path := "i=84" //"ns=8;s=plc/app/Application/sym"
	node, _ := ua.ParseNodeID(path)

	refs, _ := client.Node(node).ReferencedNodes(ctx, 0, ua.BrowseDirectionForward, ua.NodeClassAll, true)
	return len(refs) >= 1
}

// func Init(ctx context.Context, n *opcua.Node, path string) (types.Node, error) {
// 	attrs, err := n.Attributes(ctx, ua.AttributeIDNodeClass, ua.AttributeIDBrowseName, ua.AttributeIDDescription, ua.AttributeIDAccessLevel, ua.AttributeIDDataType)
// 	// fmt.Println(attrs)
// 	if err != nil {
// 		return types.Node{}, err
// 	}
//
// 	var node = types.Node{
// 		NodeID: n.ID,
// 	}
//
// 	switch err := attrs[0].Status; err {
// 	case ua.StatusOK:
// 		node.NodeClass = ua.NodeClass(attrs[0].Value.Int())
// 	default:
// 		return types.Node{}, err
// 	}
//
// 	switch err := attrs[1].Status; err {
// 	case ua.StatusOK:
// 		node.BrowseName = attrs[1].Value.String()
//
// 	default:
// 		return types.Node{}, err
// 	}
//
// 	switch err := attrs[2].Status; err {
// 	case ua.StatusOK:
// 		node.Description = attrs[2].Value.String()
// 	case ua.StatusBadAttributeIDInvalid:
// 		// ignore
// 	default:
// 		return types.Node{}, err
// 	}
//
// 	switch err := attrs[3].Status; err {
// 	case ua.StatusOK:
// 		node.AccessLevel = ua.AccessLevelType(attrs[3].Value.Int())
// 		node.Writable = node.AccessLevel&ua.AccessLevelTypeCurrentWrite == ua.AccessLevelTypeCurrentWrite
// 	case ua.StatusBadAttributeIDInvalid:
// 		// ignore
// 	default:
// 		return types.Node{}, err
// 	}
//
// 	switch err := attrs[4].Status; err {
// 	case ua.StatusOK:
// 		switch v := attrs[4].Value.NodeID().IntID(); v {
// 		case id.DateTime:
// 			node.DataType = "time.Time"
// 		case id.Boolean:
// 			node.DataType = "bool"
// 		case id.SByte:
// 			node.DataType = "int8"
// 		case id.Int16:
// 			node.DataType = "int16"
// 		case id.Int32:
// 			node.DataType = "int32"
// 		case id.Byte:
// 			node.DataType = "byte"
// 		case id.UInt16:
// 			node.DataType = "uint16"
// 		case id.UInt32:
// 			node.DataType = "uint32"
// 		case id.UtcTime:
// 			node.DataType = "time.Time"
// 		case id.String:
// 			node.DataType = "string"
// 		case id.Float:
// 			node.DataType = "float32"
// 		case id.Double:
// 			node.DataType = "float64"
// 		default:
// 			node.DataType = attrs[4].Value.NodeID().String()
// 		}
// 	case ua.StatusBadAttributeIDInvalid:
// 		// ignore
// 	default:
// 		return types.Node{}, err
// 	}
//
// 	node.Path = join(path, node.BrowseName)
// 	//fmt.Printf("%d: def.Path:%s def.NodeClass:%s\n", level, def.Path, def.NodeClass)
//
// 	return node, nil
// }
//
// func join(a, b string) string {
// 	if a == "" {
// 		return b
// 	}
// 	return a + "." + b
// }
