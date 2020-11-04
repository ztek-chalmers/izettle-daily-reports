package visma

import "izettle-daily-reports/util"

type PendingAttachment struct {
	ID          string `url:"Id,omitempty"`
	ContentType string `url:"ContentType"`
	FileName    string `url:"FileName"`
	Data        string `url:"Data,omitempty"`
	URL         string `url:"Url,omitempty"`
}

type Attachment struct {
	ID                    string    `url:"Id,omitempty"`
	ContentType           string    `url:"ContentType,omitempty"`
	DocumentID            string    `url:"DocumentId,omitempty"`
	AttachedDocumentType  int       `url:"AttachedDocumentType,omitempty"`
	FileName              string    `url:"FileName,omitempty"`
	TemporaryURL          string    `url:"TemporaryUrl,omitempty"`
	Comment               string    `url:"Comment,omitempty"`
	SupplierName          string    `url:"SupplierName,omitempty"`
	AmountInvoiceCurrency float64   `url:"AmountInvoiceCurrency,omitempty"`
	Type                  int       `url:"Type,omitempty"`
	AttachmentStatus      int       `url:"AttachmentStatus,omitempty"`
	UploadedBy            string    `url:"UploadedBy,omitempty"`
	ImageDate             util.Date `url:"ImageDate,omitempty"`
}
