# CV Analysis API

## Endpoint: POST /api/cv/analyze

Analyzes a CV using AI to extract job search parameters and generate personalized alerts.

**Access**: Pro users only (or users in trial period)

### Request

**Method**: `POST`

**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: multipart/form-data
```

**Form Data**:
- `cv` (file, required): CV file (PDF, DOCX, or TXT, max 10MB)
- `countries` (string, optional): Comma-separated list of target countries (default: "Germany")
- `locations` (string, optional): Comma-separated list of cities or "Remote" (default: "Remote")

### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "data": {
    "job_titles": [
      "Software Engineer",
      "Backend Developer",
      "Full Stack Developer"
    ],
    "skills": [
      "Go",
      "Python",
      "Docker",
      "PostgreSQL",
      "REST APIs"
    ],
    "experience_years": 5,
    "experience_level": "mid_senior",
    "languages": {
      "programming": ["Go", "Python", "JavaScript"],
      "spoken": ["en", "de"]
    },
    "exclude_keywords": [
      "junior",
      "intern",
      "graduate",
      "entry"
    ],
    "suggested_alerts": [
      {
        "keyword": "Software Engineer",
        "country": "Germany",
        "location": "Berlin, Remote",
        "local_language": "German",
        "exclude": "junior, intern, graduate, entry",
        "hours_old": 72,
        "results_wanted": 20
      }
    ],
    "target_countries": ["Germany"],
    "target_locations": ["Berlin", "Remote"]
  },
  "metadata": {
    "filename": "resume.pdf",
    "file_size_kb": 142,
    "processing_time_ms": 3241,
    "model_used": "gemini-2.5-flash"
  }
}
```

**Errors**:

- `401 Unauthorized`: Missing or invalid JWT token
- `403 Forbidden`: User is not Pro (not subscribed or trial expired)
- `400 Bad Request`: Missing CV file or invalid file format
- `500 Internal Server Error`: Analysis failed

### Example Usage

**cURL**:
```bash
curl -X POST http://localhost:8080/api/cv/analyze \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "cv=@/path/to/resume.pdf" \
  -F "countries=Germany,Netherlands" \
  -F "locations=Berlin,Amsterdam,Remote"
```

**JavaScript (Frontend)**:
```javascript
const formData = new FormData();
formData.append('cv', fileInput.files[0]);
formData.append('countries', 'Germany,Netherlands');
formData.append('locations', 'Berlin,Remote');

const response = await fetch('/api/cv/analyze', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const result = await response.json();
```

### Notes

- **Privacy**: CV files are processed in memory and immediately deleted after analysis
- **No Storage**: CVs are NOT stored in the database
- **Processing Time**: Typically 2-10 seconds depending on CV length and API response time
- **File Size Limit**: 10 MB maximum
- **Supported Formats**: PDF, DOCX, TXT
- **AI Model**: Uses Google Gemini 2.5 Flash via `jobseek-expat` CLI

### Environment Variables Required

Backend must have `GEMINI_API_KEY` set:
```bash
export GEMINI_API_KEY="your_gemini_api_key"
```
