package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type IecType uint8

// only 32 bits supported so far
const (
	BOOL  IecType = iota + 10 // 10 1 byte -> bool
	BYTE                      // 11 1 byte -> byte
	WORD                      // 12 2 byte -> uint16
	DOWRD                     // 13 4 byte -> uint32
	SINT                      // 14 1 byte -> int8
	USINT                     // 15 1 byte -> uint8
	INT                       // 16 2 byte -> int16
	UINT                      // 17 2 byte -> uint16
	DINT                      // 18 4 byte -> int32
	UDINT                     // 19 4 byte -> uint32
	REAL                      // 20 4 byte -> float32
)

type rawVar struct {
	primitive IecType
	payload   []byte
}

type IecVar struct {
	Type  string
	UUID  string
	Value any
}

func NewVar(t uint8, payload []byte) *rawVar {
	return &rawVar{
		primitive: IecType(t),
		payload:   payload,
	}
}

func (v *rawVar) Decode() (*IecVar, error) {
	buf := bytes.NewReader(v.payload)
	switch v.primitive {
	// IEC BOOL primitive
	case 10:
		var res bool
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "BOOL",
			Value: res,
		}, nil
	// IEC BYTE primitive
	case 11:
		var res byte
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "BYTE",
			Value: res,
		}, nil
	// IEC WORD primitive
	case 12:
		var res uint16
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "WORD",
			Value: res,
		}, nil
	// IEC DWORD primitive
	case 13:
		var res uint32
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "DWORD",
			Value: res,
		}, nil
	// IEC SINT primitive
	case 14:
		var res int8
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "SINT",
			Value: res,
		}, nil
	// IEC USINT primitive
	case 15:
		var res uint8
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "USINT",
			Value: res,
		}, nil
	// IEC INT primitive
	case 16:
		var res int16
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "INT",
			Value: res,
		}, nil
	// IEC UINT primitive
	case 17:
		var res uint16
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "UINT",
			Value: res,
		}, nil
	// IEC DINT primitive
	case 18:
		var res int32
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "DINT",
			Value: res,
		}, nil
	// IEC UDINT primitive
	case 19:
		var res uint32
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "UDINT",
			Value: res,
		}, nil
	// IEC REAL primitive
	case 20:
		var res float32
		if err := binary.Read(buf, binary.LittleEndian, &res); err != nil {
			return nil, err
		}
		return &IecVar{
			Type:  "REAL",
			Value: res,
		}, nil
	default:
		return nil, fmt.Errorf("invalid iec type: (%+v)", v.primitive)
	}
}
