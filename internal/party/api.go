package party

// partyDTO is the OBIE Party object on the wire. It is the single Party returned
// inside Data.Party for the public endpoints. The Address sub-object mirrors the
// OBIE PostalAddress shape, kept to the three lines the estate models.
type partyDTO struct {
	PartyID       string      `json:"PartyId"`
	PartyType     string      `json:"PartyType"`
	Name          string      `json:"Name"`
	FullLegalName string      `json:"FullLegalName,omitempty"`
	EmailAddress  string      `json:"EmailAddress,omitempty"`
	Phone         string      `json:"Phone,omitempty"`
	Address       *addressDTO `json:"Address,omitempty"`
}

// addressDTO is the OBIE PostalAddress block. AddressLine is an array per the
// OBIE schema even though the estate only carries a single line.
type addressDTO struct {
	AddressLine []string `json:"AddressLine,omitempty"`
	PostCode    string   `json:"PostCode,omitempty"`
	Country     string   `json:"Country,omitempty"`
}

// partyData is the Data block for the public party endpoints: a single Party.
type partyData struct {
	Party partyDTO `json:"Party"`
}

// toDTO converts a domain Party to its OBIE wire shape.
func toDTO(p *Party) partyDTO {
	dto := partyDTO{
		PartyID:       p.PartyID,
		PartyType:     string(p.PartyType),
		Name:          p.Name,
		FullLegalName: p.FullLegalName,
		EmailAddress:  p.EmailAddress,
		Phone:         p.Phone,
	}
	if p.AddressLine != "" || p.PostCode != "" || p.Country != "" {
		addr := &addressDTO{PostCode: p.PostCode, Country: p.Country}
		if p.AddressLine != "" {
			addr.AddressLine = []string{p.AddressLine}
		}
		dto.Address = addr
	}
	return dto
}

// internalPartyDTO is this service's own internal projection, returned to
// service-to-service callers on GET /internal/parties/{id}. It is a flat shape
// deliberately distinct from the OBIE envelope.
type internalPartyDTO struct {
	PartyID       string `json:"PartyId"`
	PartyType     string `json:"PartyType"`
	Name          string `json:"Name"`
	FullLegalName string `json:"FullLegalName,omitempty"`
	EmailAddress  string `json:"EmailAddress,omitempty"`
	Phone         string `json:"Phone,omitempty"`
	AddressLine   string `json:"AddressLine,omitempty"`
	PostCode      string `json:"PostCode,omitempty"`
	Country       string `json:"Country,omitempty"`
	Primary       bool   `json:"Primary"`
}

func toInternalDTO(p *Party) internalPartyDTO {
	return internalPartyDTO{
		PartyID:       p.PartyID,
		PartyType:     string(p.PartyType),
		Name:          p.Name,
		FullLegalName: p.FullLegalName,
		EmailAddress:  p.EmailAddress,
		Phone:         p.Phone,
		AddressLine:   p.AddressLine,
		PostCode:      p.PostCode,
		Country:       p.Country,
		Primary:       p.Primary,
	}
}
