package visma

type PendingAttachment struct {
	ID          string `url:"Id,omitempty"`
	ContentType string `url:"ContentType,omitempty"`
	FileName    string `url:"FileName,omitempty"`
	Data        string `url:"Data,omitempty"`
	URL         string `url:"Url,omitempty"`
}

type Attachment struct {
	ID                    string `url:"Id,omitempty"`
	ContentType           string `url:"ContentType,omitempty"`
	DocumentID            string `url:"DocumentId,omitempty"`
	AttachedDocumentType  int    `url:"AttachedDocumentType,omitempty"`
	FileName              string `url:"FileName,omitempty"`
	TemporaryURL          string `url:"TemporaryUrl,omitempty"`
	Comment               string `url:"Comment,omitempty"`
	SupplierName          string `url:"SupplierName,omitempty"`
	AmountInvoiceCurrency int    `url:"AmountInvoiceCurrency,omitempty"`
	Type                  int    `url:"Type,omitempty"`
	AttachmentStatus      int    `url:"AttachmentStatus,omitempty"`
	UploadedBy            string `url:"UploadedBy,omitempty"`
	ImageDate             Date   `url:"ImageDate,omitempty"`
}
