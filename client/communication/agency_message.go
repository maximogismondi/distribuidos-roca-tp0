package communication

import "strings"

const BATCH_SEPARATOR = '*'

const AGENCY_IDENTIFICATION_MESSAGE = "agency"
const BET_BATCH_MESSAGE = "bet_batch"
const FINISH_MESSAGE = "finish"
const REQUEST_RESULTS_MESSAGE = "request_results"

const MAX_BATCH_BYTES = 1024*8 - 1 // 8KB - 1 (communication delimiter)

func encodedMessage(header string, message string) string {
	return header + string(HEADER_SEPARATOR) + message
}

func EncodedIdentificationMessage(agencyId string) string {
	return encodedMessage(AGENCY_IDENTIFICATION_MESSAGE, agencyId)
}

func EncodedBetBatchMessage(batch []string) string {
	return encodedMessage(BET_BATCH_MESSAGE, strings.Join(batch, string(BATCH_SEPARATOR)))
}

func CanAppendBetToBatch(accumlatedBytes int, bet string) bool {
	return accumlatedBytes+len(bet)+1 <= MAX_BATCH_BYTES
}

func EncodedFinishMessage() string {
	return encodedMessage(FINISH_MESSAGE, "")
}

func EncodedRequestResultsMessage() string {
	return encodedMessage(REQUEST_RESULTS_MESSAGE, "")
}
