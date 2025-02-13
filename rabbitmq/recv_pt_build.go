package rabbitmq

import (
	"encoding/json"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/mitchellh/mapstructure"
)

func WrapPayloadBuild(msg []byte) {
	payloadMsg := map[string]interface{}{}
	payloadBuildMsg := agentstructs.PayloadBuildMessage{}
	err := json.Unmarshal(msg, &payloadMsg)
	if err != nil {
		logging.LogError(err, "Failed to process payload build message")
		return
	}
	payloadMsg["build_parameters"] = map[string]interface{}{
		"build_parameters": payloadMsg["build_parameters"],
	}
	err = mapstructure.Decode(&payloadMsg, &payloadBuildMsg)
	if err != nil {
		logging.LogError(err, "failed to decode message into struct")
		return
	}
	var payloadBuildResponse agentstructs.PayloadBuildResponse
	if payloadBuildFunc := agentstructs.AllPayloadData.Get(payloadBuildMsg.PayloadType).GetBuildFunction(); payloadBuildFunc == nil {
		logging.LogError(nil, "Failed to get payload build function. Do you have a function called 'build'?")
		payloadBuildResponse.Success = false
	} else {
		payloadBuildResponse = payloadBuildFunc(payloadBuildMsg)
	}
	// handle sending off the payload via a web request separately from the rest of the message
	if payloadBuildResponse.Payload != nil {
		if err := UploadPayloadData(payloadBuildMsg, payloadBuildResponse); err != nil {
			logging.LogError(err, "Failed to send payload back to Mythic via web request")
			payloadBuildResponse.BuildMessage = payloadBuildResponse.BuildMessage + "\nFailed to send payload back to Mythic: " + err.Error()
			payloadBuildResponse.Success = false
		}
	}
	if err := RabbitMQConnection.SendStructMessage(
		MYTHIC_EXCHANGE,
		PT_BUILD_RESPONSE_ROUTING_KEY,
		"",
		payloadBuildResponse,
		false,
	); err != nil {
		logging.LogError(err, "Failed to send payload response back to Mythic")
	}
	logging.LogDebug("Finished processing payload build message")

}
