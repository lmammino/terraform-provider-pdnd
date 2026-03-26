package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// DescriptorDocumentsAPI defines operations on descriptor documents and interfaces.
type DescriptorDocumentsAPI interface {
	// Documents (multiple per descriptor)
	ListDocuments(ctx context.Context, eserviceID, descriptorID uuid.UUID, offset, limit int32) ([]DescriptorDocument, Pagination, error)
	UploadDocument(ctx context.Context, eserviceID, descriptorID uuid.UUID, fileName string, fileContent []byte, prettyName string) (*DescriptorDocument, error)
	DeleteDocument(ctx context.Context, eserviceID, descriptorID uuid.UUID, documentID uuid.UUID) error
	GetDocumentByID(ctx context.Context, eserviceID, descriptorID, documentID uuid.UUID) (*DescriptorDocument, error)

	// Interface (singular, one per descriptor)
	UploadInterface(ctx context.Context, eserviceID, descriptorID uuid.UUID, fileName string, fileContent []byte, prettyName string) (*DescriptorDocument, error)
	DeleteInterface(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
	CheckInterfaceExists(ctx context.Context, eserviceID, descriptorID uuid.UUID) (bool, error)
}

type descriptorDocumentsClient struct {
	client *generated.ClientWithResponses
}

// NewDescriptorDocumentsClient creates a new DescriptorDocumentsAPI backed by the generated client.
func NewDescriptorDocumentsClient(c *generated.ClientWithResponses) DescriptorDocumentsAPI {
	return &descriptorDocumentsClient{client: c}
}

func (d *descriptorDocumentsClient) ListDocuments(ctx context.Context, eserviceID, descriptorID uuid.UUID, offset, limit int32) ([]DescriptorDocument, Pagination, error) {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)
	params := &generated.GetEServiceDescriptorDocumentsParams{Offset: offset, Limit: limit}

	resp, err := d.client.GetEServiceDescriptorDocumentsWithResponse(ctx, esID, descID, params)
	if err != nil {
		return nil, Pagination{}, fmt.Errorf("list descriptor documents: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, Pagination{}, err
	}
	if resp.JSON200 == nil {
		return nil, Pagination{}, fmt.Errorf("list descriptor documents: empty response body")
	}

	docs := make([]DescriptorDocument, len(resp.JSON200.Results))
	for i := range resp.JSON200.Results {
		converted := documentFromGenerated(&resp.JSON200.Results[i])
		docs[i] = *converted
	}
	return docs, paginationFromGenerated(resp.JSON200.Pagination), nil
}

func (d *descriptorDocumentsClient) UploadDocument(ctx context.Context, eserviceID, descriptorID uuid.UUID, fileName string, fileContent []byte, prettyName string) (*DescriptorDocument, error) {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)

	body, contentType, err := buildMultipartBody(fileName, fileContent, prettyName)
	if err != nil {
		return nil, fmt.Errorf("upload descriptor document: build multipart body: %w", err)
	}

	resp, err := d.client.UploadEServiceDescriptorDocumentWithBodyWithResponse(ctx, esID, descID, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("upload descriptor document: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return documentFromGenerated(resp.JSON201), nil
}

func (d *descriptorDocumentsClient) DeleteDocument(ctx context.Context, eserviceID, descriptorID uuid.UUID, documentID uuid.UUID) error {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)
	docID := openapi_types.UUID(documentID)

	resp, err := d.client.DeleteEServiceDescriptorDocumentWithResponse(ctx, esID, descID, docID)
	if err != nil {
		return fmt.Errorf("delete descriptor document: %w", err)
	}
	return client.CheckResponse(resp.HTTPResponse)
}

func (d *descriptorDocumentsClient) GetDocumentByID(ctx context.Context, eserviceID, descriptorID, documentID uuid.UUID) (*DescriptorDocument, error) {
	const pageSize int32 = 50
	var offset int32

	for {
		docs, pagination, err := d.ListDocuments(ctx, eserviceID, descriptorID, offset, pageSize)
		if err != nil {
			return nil, err
		}

		for i := range docs {
			if docs[i].ID == documentID {
				return &docs[i], nil
			}
		}

		offset += pageSize
		if offset >= pagination.TotalCount {
			break
		}
	}

	return nil, &client.APIError{
		StatusCode: 404,
		Title:      "Not Found",
		Detail:     fmt.Sprintf("document %s not found on descriptor %s", documentID, descriptorID),
	}
}

func (d *descriptorDocumentsClient) UploadInterface(ctx context.Context, eserviceID, descriptorID uuid.UUID, fileName string, fileContent []byte, prettyName string) (*DescriptorDocument, error) {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)

	body, contentType, err := buildMultipartBody(fileName, fileContent, prettyName)
	if err != nil {
		return nil, fmt.Errorf("upload descriptor interface: build multipart body: %w", err)
	}

	resp, err := d.client.UploadEServiceDescriptorInterfaceWithBodyWithResponse(ctx, esID, descID, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("upload descriptor interface: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return documentFromGenerated(resp.JSON201), nil
}

func (d *descriptorDocumentsClient) DeleteInterface(ctx context.Context, eserviceID, descriptorID uuid.UUID) error {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)

	resp, err := d.client.DeleteEServiceDescriptorInterfaceWithResponse(ctx, esID, descID)
	if err != nil {
		return fmt.Errorf("delete descriptor interface: %w", err)
	}
	return client.CheckResponse(resp.HTTPResponse)
}

func (d *descriptorDocumentsClient) CheckInterfaceExists(ctx context.Context, eserviceID, descriptorID uuid.UUID) (bool, error) {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)

	resp, err := d.client.DownloadEServiceDescriptorInterfaceWithResponse(ctx, esID, descID)
	if err != nil {
		return false, fmt.Errorf("check descriptor interface exists: %w", err)
	}

	switch resp.HTTPResponse.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, client.CheckResponse(resp.HTTPResponse)
	}
}

func buildMultipartBody(fileName string, fileContent []byte, prettyName string) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(fileContent); err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("prettyName", prettyName); err != nil {
		return nil, "", err
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return &buf, writer.FormDataContentType(), nil
}
