package glesys

import (
	"context"
	"encoding/json"
)

// EmailCostsDetails preserves the provider's price field types and precision.
// In production quota.amount is a JSON number, while the older model uses string.
type EmailCostsDetails struct {
	Costs     json.RawMessage `json:"costs"`
	PriceList json.RawMessage `json:"pricelist"`
}

// CostsDetails is an additive alternative to the historical fixed pricing model.
func (em *EmailDomainService) CostsDetails(ctx context.Context) (*EmailCostsDetails, error) {
	var data struct{ Response EmailCostsDetails }
	err := em.client.get(ctx, "email/costs", &data)
	return &data.Response, err
}

// EmailAccountPatch preserves omitted versus explicit zero, false and empty values.
type EmailAccountPatch struct {
	AntiSpamLevel      *int    `json:"antispamlevel,omitempty"`
	AntiVirus          *string `json:"antivirus,omitempty"`
	AutoRespond        *string `json:"autorespond,omitempty"`
	AutoRespondMessage *string `json:"autorespondmessage,omitempty"`
	QuotaInGiB         *int    `json:"quotaingib,omitempty"`
	RejectSpam         *string `json:"rejectspam,omitempty"`
}

// EditAccountPatch is an additive alternative to EditAccount.
func (em *EmailDomainService) EditAccountPatch(ctx context.Context, address string, p EmailAccountPatch) (*EmailAccount, error) {
	var data struct {
		Response struct{ EmailAccount EmailAccount }
	}
	err := em.client.post(ctx, "email/editaccount", &data, struct {
		EmailAccountPatch
		EmailAccount string `json:"emailaccount"`
	}{p, address})
	return &data.Response.EmailAccount, err
}

// CreateAccountWithSettings follows the documented API without an undocumented password argument.
func (em *EmailDomainService) CreateAccountWithSettings(ctx context.Context, address string, p EmailAccountPatch) (*EmailAccount, error) {
	var data struct {
		Response struct {
			EmailAccount EmailAccount
			Password     string `json:"password"`
		}
	}
	err := em.client.post(ctx, "email/createaccount", &data, struct {
		EmailAccountPatch
		EmailAccount string `json:"emailaccount"`
	}{p, address})
	if data.Response.EmailAccount.Password == "" {
		data.Response.EmailAccount.Password = data.Response.Password
	}
	return &data.Response.EmailAccount, err
}

// DNSRecordPatch preserves explicitly supplied empty host/data and zero TTL.
type DNSRecordPatch struct {
	RecordID int     `json:"recordid"`
	Data     *string `json:"data,omitempty"`
	Host     *string `json:"host,omitempty"`
	Type     *string `json:"type,omitempty"`
	TTL      *int    `json:"ttl,omitempty"`
}

func (s *DNSDomainService) UpdateRecordPatch(ctx context.Context, p DNSRecordPatch) (*DNSDomainRecord, error) {
	var data struct {
		Response struct{ Record DNSDomainRecord }
	}
	err := s.client.post(ctx, "domain/updaterecord", &data, p)
	return &data.Response.Record, err
}

// ServerPatch preserves explicit clearing of a description or backup schedule.
type ServerPatch struct {
	CPU         *int                    `json:"cpucores,omitempty"`
	Memory      *int                    `json:"memorysize,omitempty"`
	Storage     *int                    `json:"disksize,omitempty"`
	Hostname    *string                 `json:"hostname,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Backup      *[]ServerBackupSchedule `json:"backupschedules,omitempty"`
}

func (s *ServerService) EditPatch(ctx context.Context, id string, p ServerPatch) (*ServerDetails, error) {
	var data struct {
		Response struct{ Server ServerDetails }
	}
	err := s.client.post(ctx, "server/edit", &data, struct {
		ServerPatch
		ID string `json:"serverid"`
	}{p, id})
	return &data.Response.Server, err
}
