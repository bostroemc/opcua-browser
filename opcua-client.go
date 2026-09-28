package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/bostroemc/tui/opcua-browser/types"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type ServiceOpcUa struct {
	Client *opcua.Client
	ctx    context.Context
	config types.Config
	browse chan types.OpcUaBrowserData
	read   chan types.OpcUaReadData
	write  chan types.DataPoint
}

func NewServiceOpcUa(ctx context.Context, config types.Config, browse chan types.OpcUaBrowserData, read chan types.OpcUaReadData, write chan types.DataPoint) *ServiceOpcUa {
	return &ServiceOpcUa{
		ctx:    ctx,
		config: config,
		browse: browse,
		read:   read,
		write:  write,
	}
}

func (s *ServiceOpcUa) Connect() {
	endpoints, err := opcua.GetEndpoints(s.ctx, s.config.Server.Endpoint)
	if err != nil {
		log.Println(err)
	}
	ep, err := opcua.SelectEndpoint(endpoints, s.config.Server.Policy, ua.MessageSecurityModeFromString(s.config.Server.Mode))
	if err != nil {
		log.Println("SelectEndpoint failed: ", err)
		return
	}
	ep.EndpointURL = s.config.Server.Endpoint

	opts := []opcua.Option{
		opcua.SecurityPolicy(s.config.Server.Policy),
		opcua.SecurityModeString(s.config.Server.Mode),
		opcua.CertificateFile(s.config.Authorization.Certificate),
		opcua.PrivateKeyFile(s.config.Authorization.Key),
		opcua.AuthUsername(s.config.Authorization.Username, s.config.Authorization.Password),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeUserName),
		opcua.SessionTimeout(30 * time.Second),
	}

	s.Client, err = opcua.NewClient(ep.EndpointURL, opts...)
	if err != nil {
		log.Println(err)
	}
	if err := s.Client.Connect(s.ctx); err != nil {
		log.Println("Failed to connect to OPC UA server: ", err)
	}

	// defer s.Client.Close(s.ctx)

}

func (s *ServiceOpcUa) Run() {
	var wg sync.WaitGroup

	wg.Go(func() {
		for {
			select {
			case <-s.ctx.Done():
				return

			case a := <-s.browse:
				if isActive(s.ctx, s.Client) == false {
					time.Sleep(1 * time.Second)
				}

				refs, err := s.Client.Node(a.Node).ReferencedNodes(s.ctx, 0, ua.BrowseDirectionForward, ua.NodeClassAll, true)
				if err != nil { //TODO handle error case; check whether error occurs in case there are no referenced nodes (i.e. there are no children)
				}
				var parent types.Node
				attrs, _ := s.Client.Node(a.Node).Attributes(s.ctx, ua.AttributeIDNodeID, ua.AttributeIDBrowseName, ua.AttributeIDDescription, ua.AttributeIDAccessLevel, ua.AttributeIDDataType)
				parent = types.Node{NodeID: attrs[0].Value.NodeID(), BrowseName: attrs[1].Value.String(), Description: attrs[2].Value.String(), DataType: attrs[4].Value.String()}

				var children []types.Node
				for _, r := range refs {
					attrs, _ := r.Attributes(s.ctx, ua.AttributeIDNodeID, ua.AttributeIDBrowseName, ua.AttributeIDDescription, ua.AttributeIDAccessLevel, ua.AttributeIDDataType, ua.AttributeIDNodeClass)
					children = append(children, types.Node{NodeID: attrs[0].Value.NodeID(), BrowseName: attrs[1].Value.String(), Description: attrs[2].Value.String(), DataType: attrs[4].Value.String(), NodeClass: ua.NodeClass(attrs[5].Value.Int())})
				}

				s.browse <- types.OpcUaBrowserData{Parent: parent, Children: children}
			}
		}
	})

	//write
	wg.Go(func() {
		for {
			select {
			case <-s.ctx.Done():
				return

			case x := <-s.write:

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
				if s.Client.State() != opcua.Connected {
					continue
				}
				_, err := s.Client.Write(s.ctx, req)
				if err != nil {
					log.Println(err)
				}
			}
		}
	})

	wg.Go(func() {
		for {
			select {
			case <-s.ctx.Done():
				return

			case a := <-s.read:

				var Nodes []*ua.ReadValueID
				for _, d := range a.Data {
					_id, _ := ua.ParseNodeID(d.Node)
					Nodes = append(Nodes, &ua.ReadValueID{NodeID: _id})
				}

				if len(Nodes) > 0 {
					req := ua.ReadRequest{NodesToRead: Nodes}

					resp, err := s.Client.Read(s.ctx, &req)
					if err != nil {
						s.read <- types.OpcUaReadData{}
						continue
					}
					for i, r := range resp.Results {
						a.Data[i].Value = r.Value.Value()
					}
				}
				s.read <- a
			}
		}
	})

	wg.Wait()

	defer s.Client.Close(s.ctx)

}

func opcuaClient(ctx context.Context, config types.Config, browse chan types.OpcUaBrowserData, read chan types.OpcUaReadData, write chan types.DataPoint) {
	endpoints, err := opcua.GetEndpoints(ctx, config.Server.Endpoint)
	if err != nil {
		log.Println(err)
	}
	ep, err := opcua.SelectEndpoint(endpoints, config.Server.Policy, ua.MessageSecurityModeFromString(config.Server.Mode))
	if err != nil {
		log.Println("SelectEndpoint failed: ", err)
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
		log.Println("Failed to connect to OPC UA server: ", err)
	}

	defer c.Close(ctx)
	var wg sync.WaitGroup

	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return

			case a := <-browse:
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
					attrs, _ := s.Attributes(ctx, ua.AttributeIDNodeID, ua.AttributeIDBrowseName, ua.AttributeIDDescription, ua.AttributeIDAccessLevel, ua.AttributeIDDataType, ua.AttributeIDNodeClass)
					children = append(children, types.Node{NodeID: attrs[0].Value.NodeID(), BrowseName: attrs[1].Value.String(), Description: attrs[2].Value.String(), DataType: attrs[4].Value.String(), NodeClass: ua.NodeClass(attrs[5].Value.Int())})
				}

				browse <- types.OpcUaBrowserData{Parent: parent, Children: children}
			}
		}
	})

	//write
	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return

			case x := <-write:

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
		}
	})

	wg.Go(func() {
		var resp *ua.ReadResponse
		for {
			select {
			case <-ctx.Done():
				return

			case a := <-read:

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
		}
	})

	wg.Wait()
}

func isActive(ctx context.Context, client *opcua.Client) bool {
	path := "i=84" //"ns=8;s=plc/app/Application/sym"
	node, _ := ua.ParseNodeID(path)

	refs, _ := client.Node(node).ReferencedNodes(ctx, 0, ua.BrowseDirectionForward, ua.NodeClassAll, true)
	return len(refs) >= 1
}

type DynamicBinaryStruct struct {
	RawBytes []byte
}

func (d DynamicBinaryStruct) Encode() ([]byte, error) {
	return d.RawBytes, nil
}

func EncodeDynamicStruct(fields []*ua.StructureField, inputData map[string]interface{}) ([]byte, error) {

	buf := ua.NewBuffer(nil)

	for _, field := range fields {
		val, exists := inputData[field.Name]
		if !exists {
			return nil, fmt.Errorf("missing input data for field: %s", field.Name)
		}

		// Ensure we are working with standard Namespace 0 core types
		if field.DataType.Namespace() != 0 {
			return nil, fmt.Errorf("field %s uses nested or unhandled custom type: %s", field.Name, field.DataType.String())
		}

		// Write to the binary buffer based on the specific primitive type ID
		switch field.DataType.IntID() {
		case 1: // Boolean
			buf.WriteBool(val.(bool))
		case 2: // SByte
			buf.WriteInt8(val.(int8))
		case 3: // Byte
			buf.WriteByte(val.(byte))
		case 4: // Int16
			buf.WriteInt16(val.(int16))
		case 5: // UInt16
			buf.WriteUint16(val.(uint16))
		case 6: // Int32
			buf.WriteInt32(val.(int32))
		case 7: // UInt32
			buf.WriteUint32(val.(uint32))
		case 8: // Int64
			buf.WriteInt64(val.(int64))
		case 9: // UInt64
			buf.WriteUint64(val.(uint64))
		case 10: // Float
			buf.WriteFloat32(val.(float32))
		case 11: // Double
			buf.WriteFloat64(val.(float64))
		case 12: // String
			buf.WriteString(val.(string))
		default:
			return nil, fmt.Errorf("unsupported type ID %d for field %s", field.DataType.IntID(), field.Name)
		}

		// Check if any errors occurred during structural stream serialization
		if buf.Error() != nil {
			return nil, fmt.Errorf("buffer stream error writing field %s: %w", field.Name, buf.Error())
		}

	}
	return buf.Bytes(), nil
}

// func CallMethodWithDynamicStruct(
// 	ctx context.Context,
// 	client *opcua.Client,
// 	objectID *ua.NodeID,
// 	methodID *ua.NodeID,
// 	binaryEncodingID *ua.NodeID, // The "Default Binary" NodeID (e.g., ns=5;i=XXXX)
// 	fields []*ua.StructureField,
// 	inputData map[string]interface{},
// 	inputArguments []*ua.Variant,
// ) error {
//
// 	// 1. Generate the sequential, raw binary footprint
// 	rawBytes, err := EncodeDynamicStruct(fields, inputData)
// 	if err != nil {
// 		return fmt.Errorf("failed to encode dynamic struct: %w", err)
// 	}
//
// 	// 2. Wrap the payload manually into an unparsed Binary ExtensionObject container
// 	extObj := &ua.ExtensionObject{
// 		TypeID:       &ua.ExpandedNodeID{NodeID: binaryEncodingID}, // Links the raw payload back to its schema context
// 		EncodingMask: ua.ExtensionObjectBinary,                     // Crucial: Tells the server this is a raw byte stream
// 		Value:        DynamicBinaryStruct{RawBytes: rawBytes},      // Populates the raw structure body
// 	}
//
// 	t, _ := ua.NewVariant(extObj)
//
// 	inputArguments = append(inputArguments, t)
//
// 	req := &ua.CallMethodRequest{
// 		ObjectID:       objectID,
// 		MethodID:       methodID,
// 		InputArguments: inputArguments, // Nest the extension object into the generic Variant parameter
// 	}
//
// 	// 4. Send execution package straight to the CODESYS runtime engine
// 	resp, err := client.Call(ctx, req)
// 	if err != nil {
// 		return fmt.Errorf("network call transaction failed: %w", err)
// 	}
//
// 	if resp.StatusCode != ua.StatusOK {
// 		return fmt.Errorf("CODESYS rejected method call invocation: %s", resp.StatusCode.Error())
// 	}
//
// 	return nil
// }

// // Maps standard Namespace 0 core type NodeIDs to human-readable labels
func resolveDataType(id *ua.NodeID) string {
	if id.Namespace() != 0 {
		return "Custom Type / Structure"
	}
	switch id.IntID() {
	case 1:
		return "Boolean"
	case 2:
		return "SByte"
	case 3:
		return "Byte"
	case 4:
		return "Int16"
	case 5:
		return "UInt16"
	case 6:
		return "Int32"
	case 7:
		return "UInt32"
	case 8:
		return "Int64"
	case 9:
		return "UInt64"
	case 10:
		return "Float"
	case 11:
		return "Double"
	case 12:
		return "String"
	case 15:
		return "ByteString"
	default:
		return "Complex/Other"
	}
}

func (s *ServiceOpcUa) FindBinaryEncodingID(dataTypeID *ua.NodeID) (*ua.NodeID, error) {
	browseReq := &ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			{
				NodeID:          dataTypeID,                 // e.g., ns=5;i=3003
				BrowseDirection: ua.BrowseDirectionForward,  // Look forward from the datatype
				ReferenceTypeID: ua.NewNumericNodeID(0, 38), // i=38 is standard "HasEncoding"
				IncludeSubtypes: true,
				NodeClassMask:   0, // 0 returns all node classes
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			},
		},
	}

	resp, err := s.Client.Browse(s.ctx, browseReq)
	if err != nil {
		return nil, fmt.Errorf("network browse error: %w", err)
	}

	if len(resp.Results) == 0 || resp.Results[0].StatusCode != ua.StatusOK {
		return nil, fmt.Errorf("failed to browse encoding: status %v", resp.Results[0].StatusCode)
	}

	for _, ref := range resp.Results[0].References {
		if ref.BrowseName.Name == "Default Binary" {
			// This returns the exact *ua.NodeID needed for your ExtensionObject TypeID configuration!
			return ref.NodeID.NodeID, nil
		}
	}

	return nil, fmt.Errorf("could not find a 'Default Binary' encoding node for type %s", dataTypeID.String())
}

// func getInputArguments(ctx context.Context, client *opcua.Client, methodID *ua.NodeID) ([]*ua.ExtensionObject, error) {
func (s *ServiceOpcUa) GetInputArguments(methodID *ua.NodeID) ([]*ua.ExtensionObject, error) {
	methodNode := s.Client.Node(methodID)

	// 2. Fetch all child references of this method node
	children, err := methodNode.Children(s.ctx, 0, ua.NodeClassVariable)
	if err != nil {
		log.Fatalf("Failed to browse method child nodes: %v", err)
	}

	var inArgsNodeID *ua.NodeID

	// 3. Loop through children to look for the "InputArguments" node
	for _, child := range children {
		browseName, err := child.BrowseName(s.ctx)
		if err != nil {
			continue
		}

		// The standard OPC UA specification designates "InputArguments" for method inputs
		if browseName.Name == "InputArguments" {
			inArgsNodeID = child.ID
			break
		}
	}

	if inArgsNodeID == nil {
		log.Fatal("Could not find an InputArguments child node for this method. Does it take any parameters?")
	}

	// 4. Read the raw data variant from the discovered InputArguments node
	variantValue, err := s.Client.Node(inArgsNodeID).Value(s.ctx)
	if err != nil {
		return nil, err
	}

	// 5. Cast the underlying data into an ExtensionObject slice
	extObjects, ok := variantValue.Value().([]*ua.ExtensionObject)
	if !ok {
		return nil, nil
	}

	return extObjects, nil

}

func (s *ServiceOpcUa) GetStructureDefinition(dataType *ua.NodeID) (*ua.StructureDefinition, error) {
	attrReq := &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{
			{
				NodeID:      dataType,
				AttributeID: ua.AttributeIDDataTypeDefinition, // ID 23
			},
		},
	}
	res, err := s.Client.Read(s.ctx, attrReq)
	if err != nil || res.Results[0].Status != ua.StatusOK {
		log.Fatalf("Failed to fetch schema: %v", err)
	}

	extObj, ok := res.Results[0].Value.Value().(*ua.ExtensionObject)
	if !ok {
		fmt.Errorf("unexpected dynamic return type: %T", res.Results[0].Value.Value())
	}

	typeDef, ok := extObj.Value.(*ua.StructureDefinition)
	if !ok {
		fmt.Errorf("inner value is not a StructureDefinition: %T", extObj.Value)
	}

	return typeDef, nil
}

func getVariant(scanner *bufio.Scanner, arg *ua.Argument) (*ua.Variant, error) {
	fmt.Printf("Enter %s value for %s \n", arg.DataType, arg.Name)
	scanner.Scan()
	value := scanner.Text()
	switch arg.DataType.IntID() {
	case 1: //Boolean
		_value, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(_value)
	case 2: //SByte
		_value, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(uint8(_value))
	case 3: //Byte
		_value, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(byte(_value))
	case 4: //Int16
		_value, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(int16(_value))
	case 5: //UInt16
		_value, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(uint16(_value))
	case 6: //Int32
		_value, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(int32(_value))
	case 7: //UInt32
		_value, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(uint32(_value))
	case 8: //Int64
		_value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(_value)
	case 9: //UInt64
		_value, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(_value)
	case 10: //Float
		_value, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(float32(_value))
	case 11: //Double
		_value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(_value)
	case 12: //String
		return ua.NewVariant(value)
	case 15: //ByteString
		return ua.NewVariant([]byte(value))
	default:
		return ua.NewVariant("")
	}

}

func getExtension(scanner *bufio.Scanner, arg *ua.Argument, typeDef *ua.StructureDefinition, encoding *ua.NodeID) (*ua.Variant, error) {

	inputData := make(map[string]interface{})

	for _, field := range typeDef.Fields {
		fmt.Printf("Field Name: %s | DataType ID: %s %s\n", field.Name, field.DataType.String(), resolveDataType(field.DataType))
		fmt.Printf("Enter value for %s \n", field.Name)
		// fmt.Scanln(&value)
		scanner.Scan()
		value := scanner.Text()
		if field.DataType.String() == "i=6" {
			p, _ := strconv.ParseInt(value, 10, 32)
			inputData[field.Name] = int32(p)
		}
		if field.DataType.String() == "i=1" {
			b, _ := strconv.ParseBool(value)

			inputData[field.Name] = b
		}
		if field.DataType.String() == "i=12" {
			inputData[field.Name] = value
		}

	}

	rawBytes, err := EncodeDynamicStruct(typeDef.Fields, inputData)
	if err != nil {
		fmt.Errorf("failed to encode dynamic struct: %w", err)
	}

	// 2. Wrap the payload manually into an unparsed Binary ExtensionObject container
	extObj := &ua.ExtensionObject{
		TypeID:       &ua.ExpandedNodeID{NodeID: encoding},    // Links the raw payload back to its schema context
		EncodingMask: ua.ExtensionObjectBinary,                // Crucial: Tells the server this is a raw byte stream
		Value:        DynamicBinaryStruct{RawBytes: rawBytes}, // Populates the raw structure body
	}

	return ua.NewVariant(extObj)

}
