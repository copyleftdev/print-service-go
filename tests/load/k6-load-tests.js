import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics for detailed monitoring
const pdfGenerationRate = new Rate('pdf_generation_success');
const pdfDownloadRate = new Rate('pdf_download_success');
const minioStorageRate = new Rate('minio_storage_success');
const responseTimeTrend = new Trend('response_time_ms');
const pdfSizeCounter = new Counter('pdf_total_bytes');
const jobCounter = new Counter('jobs_processed');

// Test configuration for different scenarios
export const options = {
  scenarios: {
    // Smoke test - basic functionality
    smoke_test: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
      tags: { test_type: 'smoke' },
    },
    
    // Load test - normal expected load
    load_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 10 },  // Ramp up
        { duration: '5m', target: 10 },  // Stay at 10 users
        { duration: '2m', target: 20 },  // Ramp to 20 users
        { duration: '5m', target: 20 },  // Stay at 20 users
        { duration: '2m', target: 0 },   // Ramp down
      ],
      tags: { test_type: 'load' },
    },
    
    // Stress test - beyond normal capacity
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 10 },  // Ramp up to normal load
        { duration: '2m', target: 10 },  // Stay at normal load
        { duration: '1m', target: 25 },  // Ramp to high load
        { duration: '2m', target: 25 },  // Stay at high load
        { duration: '1m', target: 50 },  // Ramp to stress load
        { duration: '2m', target: 50 },  // Stay at stress load
        { duration: '1m', target: 0 },   // Ramp down
      ],
      tags: { test_type: 'stress' },
    },
    
    // Spike test - sudden traffic spikes
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 30 }, // Spike to 30 users
        { duration: '1m', target: 30 },  // Stay at spike
        { duration: '10s', target: 0 },   // Drop to 0
      ],
      tags: { test_type: 'spike' },
    },
    
    // Soak test - extended duration
    soak_test: {
      executor: 'constant-vus',
      vus: 20,
      duration: '30m',
      tags: { test_type: 'soak' },
    },
  },
  
  thresholds: {
    // Performance requirements
    http_req_duration: ['p(95)<2000'], // 95% of requests under 2s
    http_req_failed: ['rate<0.1'],     // Error rate under 10%
    pdf_generation_success: ['rate>0.95'], // 95% PDF generation success
    pdf_download_success: ['rate>0.95'],   // 95% PDF download success
    minio_storage_success: ['rate>0.95'],  // 95% MinIO storage success
  },
};

// Base URL for the service
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data templates for different PDF types
const testTemplates = {
  simple: {
    content: '<h1>Simple Test PDF</h1><p>This is a basic PDF generation test.</p>',
    type: 'html',
    options: { quality: 'standard' }
  },
  
  complex: {
    content: `
      <html>
        <head>
          <style>
            body { font-family: Arial, sans-serif; margin: 40px; }
            .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
            .content { margin: 20px 0; line-height: 1.6; }
            .table { width: 100%; border-collapse: collapse; margin: 20px 0; }
            .table th, .table td { border: 1px solid #ddd; padding: 12px; text-align: left; }
            .table th { background-color: #f2f2f2; }
            .footer { margin-top: 40px; font-size: 12px; color: #666; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>Complex PDF Report</h1>
            <p>Generated on: ${new Date().toISOString()}</p>
          </div>
          <div class="content">
            <h2>Executive Summary</h2>
            <p>This is a complex PDF with multiple elements including tables, styling, and formatted content.</p>
            
            <h2>Performance Metrics</h2>
            <table class="table">
              <tr><th>Metric</th><th>Value</th><th>Target</th><th>Status</th></tr>
              <tr><td>Response Time</td><td>150ms</td><td>&lt;200ms</td><td>✅ Good</td></tr>
              <tr><td>Throughput</td><td>1000 req/s</td><td>&gt;500 req/s</td><td>✅ Excellent</td></tr>
              <tr><td>Error Rate</td><td>0.1%</td><td>&lt;1%</td><td>✅ Good</td></tr>
              <tr><td>CPU Usage</td><td>45%</td><td>&lt;80%</td><td>✅ Good</td></tr>
            </table>
            
            <h2>Detailed Analysis</h2>
            <p>The system demonstrates excellent performance characteristics under various load conditions. 
            MinIO object storage integration provides reliable PDF storage and retrieval capabilities.</p>
            
            <ul>
              <li>High throughput PDF generation</li>
              <li>Reliable MinIO object storage</li>
              <li>Scalable architecture</li>
              <li>Comprehensive error handling</li>
            </ul>
          </div>
          <div class="footer">
            <p>Report ID: ${Math.random().toString(36).substr(2, 9)}</p>
            <p>Print Service Load Test - k6 Performance Validation</p>
          </div>
        </body>
      </html>
    `,
    type: 'html',
    options: { 
      quality: 'high',
      page: { size: { name: 'A4' }, orientation: 'portrait' }
    }
  },
  
  large: {
    content: generateLargeContent(),
    type: 'html',
    options: { quality: 'high' }
  },
  
  markdown: {
    content: `
# Load Test Report

## Overview
This is a **Markdown** PDF generation test for the print service.

### Features Tested
- PDF generation from Markdown
- MinIO object storage
- Concurrent processing
- Error handling

### Performance Metrics
| Metric | Value | Status |
|--------|-------|--------|
| Latency | <2s | ✅ |
| Throughput | >100 req/s | ✅ |
| Success Rate | >95% | ✅ |

### Code Example
\`\`\`javascript
const response = await fetch('/api/v1/print', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ content: markdown, type: 'markdown' })
});
\`\`\`

> This PDF was generated during k6 load testing at ${new Date().toISOString()}
    `,
    type: 'markdown',
    options: { quality: 'standard' }
  }
};

// Generate large content for stress testing
function generateLargeContent() {
  let content = '<html><body><h1>Large PDF Stress Test</h1>';
  
  for (let i = 0; i < 50; i++) {
    content += `
      <h2>Section ${i + 1}</h2>
      <p>This is section ${i + 1} of a large PDF document designed to test the service's ability to handle substantial content. `;
    
    // Add random content
    for (let j = 0; j < 20; j++) {
      content += `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. `;
    }
    
    content += '</p>';
    
    if (i % 10 === 0) {
      content += `
        <table border="1" style="width:100%; border-collapse: collapse;">
          <tr><th>Column 1</th><th>Column 2</th><th>Column 3</th></tr>
          <tr><td>Data ${i}-1</td><td>Data ${i}-2</td><td>Data ${i}-3</td></tr>
          <tr><td>Data ${i}-4</td><td>Data ${i}-5</td><td>Data ${i}-6</td></tr>
        </table>
      `;
    }
  }
  
  content += '</body></html>';
  return content;
}

// Main test function
export default function() {
  const testType = Math.random();
  let template;
  
  // Select test template based on probability
  if (testType < 0.4) {
    template = testTemplates.simple;
  } else if (testType < 0.7) {
    template = testTemplates.complex;
  } else if (testType < 0.9) {
    template = testTemplates.markdown;
  } else {
    template = testTemplates.large;
  }
  
  // Test 1: Submit PDF generation job
  const submitResponse = submitPrintJob(template);
  if (!submitResponse) return;
  
  const jobId = submitResponse.job_id;
  jobCounter.add(1);
  
  // Test 2: Poll for job completion
  const completedJob = pollJobCompletion(jobId);
  if (!completedJob) return;
  
  // Test 3: Download generated PDF
  const downloadSuccess = downloadPDF(jobId);
  
  // Test 4: Verify MinIO storage (optional health check)
  if (Math.random() < 0.1) { // 10% of requests check storage health
    checkStorageHealth();
  }
  
  // Random sleep to simulate real user behavior
  sleep(Math.random() * 2 + 0.5); // 0.5-2.5 seconds
}

// Submit a print job
function submitPrintJob(template) {
  const payload = JSON.stringify(template);
  
  const response = http.post(`${BASE_URL}/api/v1/print`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { operation: 'submit_job' },
  });
  
  const success = check(response, {
    'submit job: status is 202': (r) => r.status === 202,
    'submit job: has job_id': (r) => JSON.parse(r.body).job_id !== undefined,
    'submit job: response time < 1s': (r) => r.timings.duration < 1000,
  });
  
  pdfGenerationRate.add(success);
  responseTimeTrend.add(response.timings.duration);
  
  if (!success) {
    console.error(`Failed to submit job: ${response.status} ${response.body}`);
    return null;
  }
  
  return JSON.parse(response.body);
}

// Poll for job completion
function pollJobCompletion(jobId) {
  const maxAttempts = 30; // 30 seconds max wait
  let attempts = 0;
  
  while (attempts < maxAttempts) {
    const response = http.get(`${BASE_URL}/api/v1/jobs/${jobId}`, {
      tags: { operation: 'check_status' },
    });
    
    if (response.status === 200) {
      const job = JSON.parse(response.body);
      
      if (job.status === 'completed') {
        check(response, {
          'job completion: status is completed': () => true,
          'job completion: response time < 500ms': (r) => r.timings.duration < 500,
        });
        return job;
      } else if (job.status === 'failed') {
        console.error(`Job ${jobId} failed: ${job.error || 'Unknown error'}`);
        return null;
      }
    }
    
    attempts++;
    sleep(1); // Wait 1 second before next poll
  }
  
  console.error(`Job ${jobId} did not complete within ${maxAttempts} seconds`);
  return null;
}

// Download generated PDF
function downloadPDF(jobId) {
  const response = http.get(`${BASE_URL}/api/v1/print/${jobId}/download`, {
    tags: { operation: 'download_pdf' },
  });
  
  const success = check(response, {
    'download PDF: status is 200': (r) => r.status === 200,
    'download PDF: content-type is PDF': (r) => r.headers['Content-Type'] === 'application/pdf',
    'download PDF: has content': (r) => r.body.length > 0,
    'download PDF: response time < 2s': (r) => r.timings.duration < 2000,
  });
  
  pdfDownloadRate.add(success);
  
  if (success) {
    pdfSizeCounter.add(response.body.length);
    minioStorageRate.add(true); // PDF successfully retrieved from MinIO
  } else {
    console.error(`Failed to download PDF ${jobId}: ${response.status}`);
    minioStorageRate.add(false);
  }
  
  return success;
}

// Check storage health (MinIO connectivity)
function checkStorageHealth() {
  const response = http.get(`${BASE_URL}/health`, {
    tags: { operation: 'health_check' },
  });
  
  check(response, {
    'health check: status is 200': (r) => r.status === 200,
    'health check: response time < 100ms': (r) => r.timings.duration < 100,
  });
}

// Setup function - runs once before all tests
export function setup() {
  console.log('🚀 Starting k6 Load Tests for Print Service with MinIO');
  console.log(`📊 Target URL: ${BASE_URL}`);
  console.log('🔍 Testing scenarios: smoke, load, stress, spike, soak');
  
  // Verify service is available
  const response = http.get(`${BASE_URL}/health`);
  if (response.status !== 200) {
    throw new Error(`Service not available: ${response.status}`);
  }
  
  console.log('✅ Service health check passed');
  return { baseUrl: BASE_URL };
}

// Teardown function - runs once after all tests
export function teardown(data) {
  console.log('🏁 Load tests completed');
  console.log('📈 Check the k6 summary for detailed performance metrics');
}
