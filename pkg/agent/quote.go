package agent

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
)

type ReportData struct {
	Address         common.Address
	ContractAddress common.Address
}

func (r *ReportData) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"address":  r.Address.String(),
		"contract": r.ContractAddress.String(),
	})
}

func (r *ReportData) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{})

	binary.Write(writer, binary.BigEndian, r.Address.Bytes())
	binary.Write(writer, binary.BigEndian, r.ContractAddress.Bytes())

	return writer.Bytes(), nil
}

func generateReportDataBytes(address common.Address, contractAddress common.Address) ([]byte, error) {
	reportData := &ReportData{
		Address:         address,
		ContractAddress: contractAddress,
	}

	return reportData.MarshalBinary()
}
