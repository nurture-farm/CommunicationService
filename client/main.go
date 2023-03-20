/*
 *  Copyright 2023 NURTURE AGTECH PVT LTD
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package main

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	common "github.com/nurture-farm/Contracts/Common/Gen/GoCommon"
	commEngine "github.com/nurture-farm/Contracts/CommunicationEngine/Gen/GoCommunicationEngine"
)

const (
	LOCAL_URL = ":8020"
	DEV_URL   = "internal-a2d376e6948514a73a514e3584160823-409853754.ap-south-1.elb.amazonaws.com:80"
	STAGE_URL = "internal-a49872e2b408d43d28ce6000e575e84b-437549400.ap-south-1.elb.amazonaws.com:80"
	PROD_URL  = "internal-a85f3a2f550f84f578459537c8ae3304-113299092.ap-south-1.elb.amazonaws.com:80"
)

var ENV = PROD_URL

func main() {
	TestSendCommunication()
	//TestSendBulkCommunication()
}

func TestSendCommunication() {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(ENV, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := commEngine.NewCommunicationEngineClient(conn)

	request := &commEngine.CommunicationEvent{
		ReceiverActorDetails: &commEngine.ActorDetails{
			MobileNumber: "9453849441",
			LanguageCode: common.LanguageCode_EN_US,
		},
		TemplateName: "farmer_booking_reject_eng",
		Channel: []common.CommunicationChannel{
			common.CommunicationChannel_SMS,
		},
		//Media: &commEngine.Media{
		//	MediaType:       common.MediaType_DOCUMENT,
		//	MediaAccessType: common.MediaAccessType_PUBLIC_URL,
		//	MediaInfo:       "https://awd-dsr-signup-media-files.s3.ap-south-1.amazonaws.com/AWD+Hindi+Leaflet+digital+(1).pdf",
		//	DocumentName:    "AWD method for Paddy",
		//},
	}

	response, err := c.SendCommunication(context.Background(), request)
	if err != nil {
		log.Fatalf("Error when calling SendCommunication: %s", err)
	}
	log.Println(response)
}

func TestSendBulkCommunication() {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(ENV, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := commEngine.NewCommunicationEngineClient(conn)

	request := &commEngine.BulkCommunicationEvent{
		CommunicationEvents: []*commEngine.CommunicationEvent{
			{
				ReceiverActorDetails: &commEngine.ActorDetails{
					MobileNumber: "7086880268",
					LanguageCode: common.LanguageCode_EN_US,
				},
				TemplateName: "farmer_booking_reject",
				Channel: []common.CommunicationChannel{
					common.CommunicationChannel_SMS,
				},
			},
			{
				ReceiverActorDetails: &commEngine.ActorDetails{
					MobileNumber: "9453849441",
					LanguageCode: common.LanguageCode_EN_US,
				},
				TemplateName: "farmer_booking_reject",
				Channel: []common.CommunicationChannel{
					common.CommunicationChannel_SMS,
				},
			},
		},
	}

	response, err := c.SendBulkCommunication(context.Background(), request)
	if err != nil {
		log.Fatalf("Error when calling SendBulkCommunication: %s", err)
	}
	log.Println(response)
}
