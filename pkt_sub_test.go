/*
 * Copyright Go-IIoT (https://github.com/goiiot)
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

package libmqtt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	std "github.com/eclipse/paho.mqtt.golang/packets"
)

func TestSubAck5(t *testing.T) {

	timeStr := "1234567890"

	// 90
	// 1b
	// 0000
	// 18
	// 26 0009 756e69782d74696d65 000a 31323334353637383930

	// write an ack
	suback := &SubAckPacket{
		Props: &SubAckProps{
			// Reason:    "ok",
			UserProps: UserProps{"unix-time": []string{timeStr}},
		},
	}
	suback.SetVersion(5)

	var buff bytes.Buffer

	err := suback.WriteTo(&buff)
	if err != nil {
		t.Error("got error", err)
	}
	derivedHex := hex.EncodeToString(buff.Bytes())
	fmt.Println("derived hex", derivedHex) // 901b000018260009756e69782d74696d65000a31323334353637383930

	hexdata1 := derivedHex
	data, _ := hex.DecodeString(hexdata1)
	r := bytes.NewReader(data)

	p, err := Decode(5, r)
	if err != nil {
		t.Error("got error", err)
	}
	fmt.Println("got packet", p, err)
	sub := p.(*SubAckPacket)
	got, ok := sub.Props.UserProps.Get("unix-time")
	if got != timeStr || !ok {
		t.Error("wanted ok, got", sub.Props.Reason)
	}

}

func TestSubbasic5(t *testing.T) {

	// These hex strings were glommed from 'pip3 install gmqtt'
	// which seems to know mqtt5 better

	hexdata1 := "8213000100000d544553542f54494d456162636400"
	data, _ := hex.DecodeString(hexdata1)
	r := bytes.NewReader(data)

	p, err := Decode(5, r)
	fmt.Println("got packet", p, err)
	sub := p.(*SubscribePacket)
	if sub.Topics[0].Name != "TEST/TIMEabcd" {
		t.Error("wanted TEST/TIMEabcd, got", sub.Topics[0])
	}

	hexdata2 := "823000021a2600046b657931000476616c312600046b657932000476616c320010544553542f54494d4565666768696a6b01"
	data, _ = hex.DecodeString(hexdata2)
	r = bytes.NewReader(data)
	p, err = Decode(5, r)
	fmt.Println("got packet", p, err)
	sub = p.(*SubscribePacket)
	if sub.Topics[0].Name != "TEST/TIMEefghijk" {
		t.Error("wanted TEST/TIMEefghijk, got", sub.Topics[0])
	}
	val, _ := sub.Props.UserProps.Get("key1")
	fmt.Println("got val", val)
	val, _ = sub.Props.UserProps.Get("key2")
	fmt.Println("got val", val)
	if val != "val2" {
		t.Error("wanted val2")
	}
	fmt.Println("got PacketID", sub.PacketID)
	fmt.Println("got Props.SubID", sub.Props.SubID)

	// now go the other way.
	sub = &SubscribePacket{
		Topics: []*Topic{
			{Name: "TEST/TIMEefghijk", Qos: 1},
		},
		PacketID: 2,
		Props: &SubscribeProps{
			SubID:     0,
			UserProps: UserProps{"key1": []string{"val1"}, "key2": []string{"val2"}},
		},
	}
	sub.SetVersion(5)

	var buff bytes.Buffer

	err = sub.WriteTo(&buff)
	_ = err
	derivedHex := hex.EncodeToString(buff.Bytes())
	if derivedHex != hexdata2 {
		t.Error("our result must match the gmqtt serialization")
	}

}

// sub test data
var (
	testSubTopics   []*Topic
	testSubMsgs     []*SubscribePacket
	testSubAckMsgs  []*SubAckPacket
	testUnSubMsgs   []*UnsubPacket
	testUnSubAckMsg *UnsubAckPacket

	// mqtt 3.1.1
	testSubMsgBytesV311      [][]byte
	testSubAckMsgBytesV311   [][]byte
	testUnSubMsgBytesV311    [][]byte
	testUnSubAckMsgBytesV311 []byte

	// mqtt 5.0
	testSubMsgBytesV5      [][]byte
	testSubAckMsgBytesV5   [][]byte
	testUnSubMsgBytesV5    [][]byte
	testUnSubAckMsgBytesV5 []byte
)

func initTestData_Sub() {
	size := len(testTopics)
	testSubTopics = make([]*Topic, size)
	for i := range testSubTopics {
		testSubTopics[i] = &Topic{Name: testTopics[i], Qos: testTopicQos[i]}
	}

	testSubMsgs = make([]*SubscribePacket, size)
	testSubAckMsgs = make([]*SubAckPacket, size)
	testUnSubMsgs = make([]*UnsubPacket, size)

	testSubMsgBytesV311 = make([][]byte, size)
	testSubAckMsgBytesV311 = make([][]byte, size)
	testUnSubMsgBytesV311 = make([][]byte, size)

	testSubMsgBytesV5 = make([][]byte, size)
	testSubAckMsgBytesV5 = make([][]byte, size)
	testUnSubMsgBytesV5 = make([][]byte, size)

	for i := range testTopics {
		msgID := uint16(i + 1)
		testSubMsgs[i] = &SubscribePacket{
			Topics:   testSubTopics[:i+1],
			PacketID: msgID,
			Props: &SubscribeProps{
				SubID:     100,
				UserProps: testConstUserProps,
			},
		}

		subPkt := std.NewControlPacket(std.Subscribe).(*std.SubscribePacket)
		subPkt.Topics = testTopics[:i+1]
		subPkt.Qoss = testTopicQos[:i+1]
		subPkt.MessageID = msgID

		subBuf := new(bytes.Buffer)
		_ = subPkt.Write(subBuf)
		testSubMsgBytesV311[i] = subBuf.Bytes()
		testSubMsgBytesV5[i] = newV5TestPacketBytes(CtrlSubscribe, 0, nil, nil)

		testSubAckMsgs[i] = &SubAckPacket{
			PacketID: msgID,
			Codes:    testSubAckCodes[:i+1],
			Props: &SubAckProps{
				Reason:    "MQTT",
				UserProps: testConstUserProps,
			},
		}
		subAckPkt := std.NewControlPacket(std.Suback).(*std.SubackPacket)
		subAckPkt.MessageID = msgID
		subAckPkt.ReturnCodes = testSubAckCodes[:i+1]
		subAckBuf := new(bytes.Buffer)
		_ = subAckPkt.Write(subAckBuf)
		testSubAckMsgBytesV311[i] = subAckBuf.Bytes()
		testSubAckMsgBytesV5[i] = newV5TestPacketBytes(CtrlSubAck, 0, nil, nil)

		testUnSubMsgs[i] = &UnsubPacket{
			PacketID:   msgID,
			TopicNames: testTopics[:i+1],
			Props: &UnsubProps{
				UserProps: testConstUserProps,
			},
		}
		unsubPkt := std.NewControlPacket(std.Unsubscribe).(*std.UnsubscribePacket)
		unsubPkt.Topics = testTopics[:i+1]
		unsubPkt.MessageID = msgID
		unSubBuf := new(bytes.Buffer)
		_ = unsubPkt.Write(unSubBuf)
		testUnSubMsgBytesV311[i] = unSubBuf.Bytes()
		testUnSubMsgBytesV5[i] = newV5TestPacketBytes(CtrlUnSub, 0, nil, nil)
	}

	unSunAckBuf := new(bytes.Buffer)
	testUnSubAckMsg = &UnsubAckPacket{
		PacketID: 1,
		Props: &UnsubAckProps{
			Reason:    "MQTT",
			UserProps: testConstUserProps,
		},
	}
	unsubAckPkt := std.NewControlPacket(std.Unsuback).(*std.UnsubackPacket)
	unsubAckPkt.MessageID = 1
	_ = unsubAckPkt.Write(unSunAckBuf)
	testUnSubAckMsgBytesV311 = unSunAckBuf.Bytes()
	testUnSubAckMsgBytesV5 = newV5TestPacketBytes(CtrlUnSubAck, 0, nil, nil)
}

func TestSubscribePacket_Bytes(t *testing.T) {
	for i, p := range testSubMsgs {
		testPacketBytes(V311, p, testSubMsgBytesV311[i], t)
		//testPacketBytes(V5, p, testSubMsgBytesV5[i], t)
	}
}

func TestSubAckPacket_Bytes(t *testing.T) {
	for i, p := range testSubAckMsgs {
		testPacketBytes(V311, p, testSubAckMsgBytesV311[i], t)
		// atw todo: finish these. testPacketBytes(V5, p, testSubAckMsgBytesV5[i], t)
	}
}

func TestUnSubPacket_Bytes(t *testing.T) {
	for i, p := range testUnSubMsgs {
		testPacketBytes(V311, p, testUnSubMsgBytesV311[i], t)
		//testPacketBytes(V5, p, testUnSubMsgBytesV5[i], t)
	}
}

func TestUnSubAckPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testUnSubAckMsg, testUnSubAckMsgBytesV311, t)

	t.Skip("v5")
	testPacketBytes(V5, testUnSubAckMsg, testUnSubAckMsgBytesV5, t)
}
