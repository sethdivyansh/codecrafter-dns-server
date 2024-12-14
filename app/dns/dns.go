package dns

import (
	"encoding/binary"
	"fmt"
)

type HeaderFlags struct {
	QR     uint16
	OPCODE uint16
	AA     uint16
	TC     uint16
	RD     uint16
	RA     uint16
	Z      uint16
	RCODE  uint16
}

type Header struct {
	ID      uint16
	Flags   HeaderFlags
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type Question struct {
	Name  []byte
	Type  uint16
	Class uint16
}

type Answer struct {
	Name   []byte
	Type   uint16
	Class  uint16
	TTL    uint32
	Length uint16
	Data   uint32
}

type Message struct {
	Header   []byte
	Question []byte
	Answer   []byte
}

func NewMessage() *Message {
	return &Message{
		Header:   make([]byte, 12),
		Question: []byte{},
		Answer:   []byte{},
	}
}

func (m *Message) SetHeader(header Header) {
	flags := combineFlags(uint(header.Flags.QR), uint(header.Flags.OPCODE), uint(header.Flags.AA), uint(header.Flags.TC), uint(header.Flags.RD), uint(header.Flags.RA), uint(header.Flags.Z), uint(header.Flags.RCODE))

	binary.BigEndian.PutUint16(m.Header[0:2], header.ID)
	binary.BigEndian.PutUint16(m.Header[2:4], flags)
	binary.BigEndian.PutUint16(m.Header[4:6], header.QDCount)
	binary.BigEndian.PutUint16(m.Header[6:8], header.ANCount)
	binary.BigEndian.PutUint16(m.Header[8:10], header.NSCount)
	binary.BigEndian.PutUint16(m.Header[10:12], header.ARCount)
}

func (m *Message) SetQuestion(question Question) {
	m.Question = append(m.Question, question.Name...)
	m.Question = binary.BigEndian.AppendUint16(m.Question, question.Type)
	m.Question = binary.BigEndian.AppendUint16(m.Question, question.Class)
}

func (m *Message) SetAnswer(answer Answer) {
	m.Answer = append(m.Answer, answer.Name...)
	m.Answer = binary.BigEndian.AppendUint16(m.Answer, answer.Type)
	m.Answer = binary.BigEndian.AppendUint16(m.Answer, answer.Class)
	m.Answer = binary.BigEndian.AppendUint32(m.Answer, answer.TTL)
	m.Answer = binary.BigEndian.AppendUint16(m.Answer, answer.Length)
	m.Answer = binary.BigEndian.AppendUint32(m.Answer, answer.Data)
}

func PrepareMessage(receivedData *[]byte) *Message {
	// qnaSection := (*receivedData)[12:]

	// Find the end of the domain name (question name section)
	qnameEnd := 12
	for ; (*receivedData)[qnameEnd] != 0; qnameEnd++ {
	}

	header := createHeader((*receivedData)[:12])
	// question := createQuestion((*receivedData)[12 : qnameEnd+5])
	questions := createQuestions((*receivedData)[12 : ], header.QDCount)

	// answers := []Answer{}
	message := NewMessage()
	message.SetHeader(header)
	// Create answer based on question
	answers := []Answer{}
	for _, question := range questions {
		fmt.Println("Question:", question)
		ques := createQuestion(question)
		answer := Answer{
			Name:   ques.Name,
			Type:   ques.Type,
			Class:  ques.Class,
			TTL:    60,
			Length: 4,
			Data:   binary.BigEndian.Uint32([]byte{8, 8, 8, 8}), // Example: 8.8.8.8
		}

		fmt.Println("Answer:", answer)
		answers = append(answers, answer)

		message.SetQuestion(ques)
		
	}

	for _, answer := range answers {
		message.SetAnswer(answer)
	}
	

	return message
}

func createHeader(header []byte) Header{
	var rcode uint
	fmt.Println("Header Slice:", header)
	opcode := (header[2] >> 3) & 0x0F

	if opcode != 0 {
		rcode = 4
	} else {
		rcode = 0
	}
	return Header{
		ID: binary.BigEndian.Uint16(header[0:2]),
		Flags: HeaderFlags{
			QR:     1,
			OPCODE: uint16((header[2] >> 3) & 0x0F),
			AA:     0,
			// AA:     uint16(header[2] >> 2 & 0x01),
			// TC:     uint16(header[2] >> 1 & 0x01),
			TC:     0,
			RD:     uint16(header[2] & 0x01),
			// RA:     uint16(header[3] >> 7),
			// Z:      uint16((header[3] >> 4) & 0x07),
			RA:     0,
			Z:      0,
			RCODE:  uint16(rcode),
		},
		QDCount: binary.BigEndian.Uint16(header[4:6]),
		ANCount: binary.BigEndian.Uint16(header[6:8]),
		NSCount: binary.BigEndian.Uint16(header[8:10]),
		ARCount: binary.BigEndian.Uint16(header[10:12]),
	}
}

func createQuestions(questionSection []byte, numQues uint16) []byte {
	var questions []byte
	qnameEnd := 0
	fmt.Println("Question Section:", questionSection)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in createQuestions:", r)
		}
	}()
	for i := 0; i < int(numQues); i++ {
		fmt.Println("Inside Loop")
		for ; questionSection[qnameEnd] != 0; {
			qnameEnd++
		}
		// question := createQuestion(questionSection[:qnameEnd+5])
		question := questionSection[:qnameEnd+5]
		fmt.Println("QuestionCreated:", question)
		questions = append(questions, question...)
		questionSection = questionSection[qnameEnd+5:]
		qnameEnd += 5
	}
	fmt.Println("Questions:", questions)
	return questions
}

func createQuestion(question []byte) Question {
	fmt.Println("Question create:", question)
	name := question[:len(question)-4]
	qtype := binary.BigEndian.Uint16(question[len(question)-4 : len(question)-2])
	qclass := binary.BigEndian.Uint16(question[len(question)-2:])

	return Question{
		Name:  name,
		Type:  qtype,
		Class: qclass,
	}
}

// func createAnswer(question Question) Answer {
// 	// In this case, we'll return the IP address 8.8.8.8 as an example
// 	ipAddress := binary.BigEndian.Uint32([]byte{8, 8, 8, 8})

// 	return Answer{
// 		Name:   question.Name, // Copy the domain name from the question
// 		Type:   question.Type, // Should be 1 (A record)
// 		Class:  question.Class, // Should be 1 (IN)
// 		TTL:    60, // Time-to-live for the response
// 		Length: 4, // IPv4 addresses are 4 bytes long
// 		Data:   ipAddress, // The IP address for the answer (8.8.8.8 as an example)
// 	}
// }

func combineFlags(qr, opcode, aa, tc, rd, ra, z, rcode uint) uint16 {
	return uint16((qr << 15) | (opcode << 11) | (aa << 10) | (tc << 9) | (rd << 8) | (ra << 7) | (z << 4) | rcode)
}