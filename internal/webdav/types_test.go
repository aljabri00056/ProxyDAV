package webdav

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestMultistatus_XMLMarshaling(t *testing.T) {
	multistatus := Multistatus{
		Responses: []Response{
			{
				Href: "/test/file.txt",
				Propstat: Propstat{
					Prop: Prop{
						DisplayName:   "file.txt",
						ContentLength: func() *int64 { i := int64(1024); return &i }(),
						ContentType:   "text/plain",
						LastModified:  "Mon, 01 Jan 2024 12:00:00 GMT",
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
		},
	}

	data, err := xml.MarshalIndent(multistatus, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal multistatus: %v", err)
	}

	xmlStr := string(data)

	// Check for expected XML elements (accounting for namespace attributes)
	expectedElements := []string{
		`multistatus`,
		`response`,
		`<href>/test/file.txt</href>`,
		`propstat`,
		`prop`,
		`<displayname>file.txt</displayname>`,
		`<getcontentlength>1024</getcontentlength>`,
		`<getcontenttype>text/plain</getcontenttype>`,
		`<getlastmodified>Mon, 01 Jan 2024 12:00:00 GMT</getlastmodified>`,
		`<status>HTTP/1.1 200 OK</status>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(xmlStr, expected) {
			t.Errorf("Expected XML to contain %s, but it didn't. XML: %s", expected, xmlStr)
		}
	}
}

func TestMultistatus_XMLUnmarshaling(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<multistatus xmlns="DAV:">
  <response>
    <href>/test/file.txt</href>
    <propstat>
      <prop>
        <displayname>file.txt</displayname>
        <getcontentlength>1024</getcontentlength>
        <getcontenttype>text/plain</getcontenttype>
        <getlastmodified>Mon, 01 Jan 2024 12:00:00 GMT</getlastmodified>
      </prop>
      <status>HTTP/1.1 200 OK</status>
    </propstat>
  </response>
</multistatus>`

	var multistatus Multistatus
	err := xml.Unmarshal([]byte(xmlData), &multistatus)
	if err != nil {
		t.Fatalf("Failed to unmarshal multistatus: %v", err)
	}

	if len(multistatus.Responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(multistatus.Responses))
	}

	response := multistatus.Responses[0]
	if response.Href != "/test/file.txt" {
		t.Errorf("Expected href '/test/file.txt', got '%s'", response.Href)
	}

	prop := response.Propstat.Prop
	if prop.DisplayName != "file.txt" {
		t.Errorf("Expected display name 'file.txt', got '%s'", prop.DisplayName)
	}

	if prop.ContentLength == nil || *prop.ContentLength != 1024 {
		t.Errorf("Expected content length 1024, got %v", prop.ContentLength)
	}

	if prop.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got '%s'", prop.ContentType)
	}

	if prop.LastModified != "Mon, 01 Jan 2024 12:00:00 GMT" {
		t.Errorf("Expected last modified 'Mon, 01 Jan 2024 12:00:00 GMT', got '%s'", prop.LastModified)
	}

	if response.Propstat.Status != "HTTP/1.1 200 OK" {
		t.Errorf("Expected status 'HTTP/1.1 200 OK', got '%s'", response.Propstat.Status)
	}
}

func TestResponse_DirectoryXML(t *testing.T) {
	response := Response{
		Href: "/documents/",
		Propstat: Propstat{
			Prop: Prop{
				DisplayName:  "documents",
				ResourceType: &ResourceType{Collection: &Collection{}},
			},
			Status: "HTTP/1.1 200 OK",
		},
	}

	data, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal directory response: %v", err)
	}

	xmlStr := string(data)

	// Check for directory-specific elements (accounting for namespace attributes)
	expectedElements := []string{
		`response`,
		`<href>/documents/</href>`,
		`<displayname>documents</displayname>`,
		`resourcetype`,
		`collection`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(xmlStr, expected) {
			t.Errorf("Expected XML to contain %s, but it didn't. XML: %s", expected, xmlStr)
		}
	}
}

func TestPropFind_XMLMarshaling(t *testing.T) {
	// Test PropFind with specific properties
	propFind := PropFind{
		Prop: &PropReq{
			DisplayName:   &struct{}{},
			ResourceType:  &struct{}{},
			ContentLength: &struct{}{},
			ContentType:   &struct{}{},
			LastModified:  &struct{}{},
		},
	}

	data, err := xml.MarshalIndent(propFind, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal PropFind: %v", err)
	}

	xmlStr := string(data)

	expectedElements := []string{
		`propfind`,
		`prop`,
		`<displayname></displayname>`,
		`<resourcetype></resourcetype>`,
		`<getcontentlength></getcontentlength>`,
		`<getcontenttype></getcontenttype>`,
		`<getlastmodified></getlastmodified>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(xmlStr, expected) {
			t.Errorf("Expected XML to contain %s, but it didn't. XML: %s", expected, xmlStr)
		}
	}
}

func TestPropFind_AllProp(t *testing.T) {
	// Test PropFind with allprop
	propFind := PropFind{
		AllProp: &struct{}{},
	}

	data, err := xml.MarshalIndent(propFind, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal PropFind with allprop: %v", err)
	}

	xmlStr := string(data)

	expectedElements := []string{
		`propfind`,
		`<allprop></allprop>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(xmlStr, expected) {
			t.Errorf("Expected XML to contain %s, but it didn't. XML: %s", expected, xmlStr)
		}
	}
}

func TestPropFind_XMLUnmarshaling(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<propfind xmlns="DAV:">
  <prop>
    <displayname/>
    <resourcetype/>
    <getcontentlength/>
    <getcontenttype/>
    <getlastmodified/>
    <creationdate/>
    <getetag/>
  </prop>
</propfind>`

	var propFind PropFind
	err := xml.Unmarshal([]byte(xmlData), &propFind)
	if err != nil {
		t.Fatalf("Failed to unmarshal PropFind: %v", err)
	}

	if propFind.Prop == nil {
		t.Fatal("Expected Prop to be set, got nil")
	}

	if propFind.AllProp != nil {
		t.Error("Expected AllProp to be nil when Prop is set")
	}

	prop := propFind.Prop
	if prop.DisplayName == nil {
		t.Error("Expected DisplayName to be set")
	}
	if prop.ResourceType == nil {
		t.Error("Expected ResourceType to be set")
	}
	if prop.ContentLength == nil {
		t.Error("Expected ContentLength to be set")
	}
	if prop.ContentType == nil {
		t.Error("Expected ContentType to be set")
	}
	if prop.LastModified == nil {
		t.Error("Expected LastModified to be set")
	}
	if prop.CreationDate == nil {
		t.Error("Expected CreationDate to be set")
	}
	if prop.ETag == nil {
		t.Error("Expected ETag to be set")
	}
}

func TestFormatTime(t *testing.T) {
	// Test with a specific time
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	expected := "Mon, 01 Jan 2024 12:00:00 GMT"
	result := FormatTime(testTime)

	if result != expected {
		t.Errorf("Expected formatted time '%s', got '%s'", expected, result)
	}

	// Test with different timezone (should convert to UTC)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("Skipping timezone test: %v", err)
	}

	testTimeNY := time.Date(2024, 1, 1, 7, 0, 0, 0, loc) // 7 AM EST = 12 PM UTC
	result = FormatTime(testTimeNY)

	if result != expected {
		t.Errorf("Expected formatted time '%s' (converted to UTC), got '%s'", expected, result)
	}
}

func TestGenerateETag(t *testing.T) {
	url := "https://example.com/file.txt"
	modTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	expected := `"https://example.com/file.txt-20240101120000"`
	result := GenerateETag(url, modTime)

	if result != expected {
		t.Errorf("Expected ETag '%s', got '%s'", expected, result)
	}

	// Test with different URL and time
	url2 := "https://test.com/document.pdf"
	modTime2 := time.Date(2024, 12, 25, 15, 30, 45, 0, time.UTC)

	expected2 := `"https://test.com/document.pdf-20241225153045"`
	result2 := GenerateETag(url2, modTime2)

	if result2 != expected2 {
		t.Errorf("Expected ETag '%s', got '%s'", expected2, result2)
	}

	// Test that different inputs produce different ETags
	if result == result2 {
		t.Error("Different inputs should produce different ETags")
	}
}

func TestComplexMultistatus(t *testing.T) {
	// Test with multiple responses including both files and directories
	multistatus := Multistatus{
		Responses: []Response{
			{
				Href: "/documents/",
				Propstat: Propstat{
					Prop: Prop{
						DisplayName:  "documents",
						ResourceType: &ResourceType{Collection: &Collection{}},
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
			{
				Href: "/documents/file1.txt",
				Propstat: Propstat{
					Prop: Prop{
						DisplayName:   "file1.txt",
						ContentLength: func() *int64 { i := int64(2048); return &i }(),
						ContentType:   "text/plain",
						LastModified:  "Tue, 02 Jan 2024 14:30:00 GMT",
						ETag:          `"file1-20240102143000"`,
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
			{
				Href: "/documents/image.png",
				Propstat: Propstat{
					Prop: Prop{
						DisplayName:   "image.png",
						ContentLength: func() *int64 { i := int64(512000); return &i }(),
						ContentType:   "image/png",
						LastModified:  "Wed, 03 Jan 2024 09:15:30 GMT",
						ETag:          `"image-20240103091530"`,
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
		},
	}

	data, err := xml.MarshalIndent(multistatus, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal complex multistatus: %v", err)
	}

	// Unmarshal to verify roundtrip
	var unmarshaled Multistatus
	err = xml.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal complex multistatus: %v", err)
	}

	if len(unmarshaled.Responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(unmarshaled.Responses))
	}

	// Check directory response
	dirResponse := unmarshaled.Responses[0]
	if dirResponse.Href != "/documents/" {
		t.Errorf("Expected directory href '/documents/', got '%s'", dirResponse.Href)
	}
	if dirResponse.Propstat.Prop.ResourceType == nil || dirResponse.Propstat.Prop.ResourceType.Collection == nil {
		t.Error("Expected directory to have collection resource type")
	}

	// Check file responses
	fileResponse1 := unmarshaled.Responses[1]
	if fileResponse1.Propstat.Prop.ContentLength == nil || *fileResponse1.Propstat.Prop.ContentLength != 2048 {
		t.Errorf("Expected file1 content length 2048, got %v", fileResponse1.Propstat.Prop.ContentLength)
	}

	fileResponse2 := unmarshaled.Responses[2]
	if fileResponse2.Propstat.Prop.ContentType != "image/png" {
		t.Errorf("Expected file2 content type 'image/png', got '%s'", fileResponse2.Propstat.Prop.ContentType)
	}
}
