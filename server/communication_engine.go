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

package server

import (
	"context"
	"github.com/google/uuid"
	"github.com/nurture-farm/CommunicationService/kafka"
	commEngine "github.com/nurture-farm/Contracts/CommunicationEngine/Gen/GoCommunicationEngine"
)

type CommunicationEngineServer struct {
}

var CommEngine *CommunicationEngineServer = &CommunicationEngineServer{}

func (s *CommunicationEngineServer) SendCommunication(ctx context.Context, event *commEngine.CommunicationEvent) (*commEngine.CommunicationResponse, error) {

	ceResponse, err := pushToKafka(event)
	if err != nil {
		return nil, err
	}
	return ceResponse, nil
}

func (s *CommunicationEngineServer) SendBulkCommunication(ctx context.Context, events *commEngine.BulkCommunicationEvent) (*commEngine.BulkCommunicationResponse, error) {

	response := &commEngine.BulkCommunicationResponse{}
	for _, event := range events.CommunicationEvents {
		ceResponse, err := pushToKafka(event)
		if err != nil {
			return nil, err
		}
		response.CommunicationResponses = append(response.CommunicationResponses, ceResponse)
	}
	return response, nil
}

func pushToKafka(event *commEngine.CommunicationEvent) (*commEngine.CommunicationResponse, error) {

	var refID string
	if event.ReferenceId == "" {
		refID = uuid.New().String()
	} else {
		refID = event.ReferenceId
	}
	event.ReferenceId = refID
	if err := kafka.SendMessage(event); err != nil {
		return nil, err
	}
	ceResponse := &commEngine.CommunicationResponse{
		ReferenceId: refID,
	}
	return ceResponse, nil
}
