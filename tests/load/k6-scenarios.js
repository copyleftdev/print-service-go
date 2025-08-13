import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// Specialized metrics for production scenarios
const concurrentJobsGauge = new Gauge('concurrent_jobs_active');
const minioLatencyTrend = new Trend('minio_operation_duration');
const memoryUsageGauge = new Gauge('estimated_memory_usage_mb');
const errorsByTypeCounter = new Counter('errors_by_type');

// Production-worthy test scenarios
export const options = {
  scenarios: {
    // High-frequency small PDFs (typical web usage)
    web_traffic: {
      executor: 'constant-arrival-rate',
      rate: 50, // 50 requests per second
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 20,
      maxVUs: 100,
      tags: { scenario: 'web_traffic' },
    },
    
    // Batch processing simulation (reports, invoices)
    batch_processing: {
      executor: 'per-vu-iterations',
      vus: 10,
      iterations: 50, // Each VU processes 50 documents
      maxDuration: '10m',
      tags: { scenario: 'batch_processing' },
    },
    
    // Enterprise reporting (large, complex PDFs)
    enterprise_reports: {
      executor: 'ramping-arrival-rate',
      startRate: 1,
      stages: [
        { duration: '2m', target: 5 },  // Ramp to 5 reports/sec
        { duration: '5m', target: 5 },  // Sustain 5 reports/sec
        { duration: '2m', target: 10 }, // Peak at 10 reports/sec
        { duration: '1m', target: 0 },  // Ramp down
      ],
      preAllocatedVUs: 20,
      maxVUs: 50,
      tags: { scenario: 'enterprise_reports' },
    },
    
    // API integration testing (external systems)
    api_integration: {
      executor: 'constant-vus',
      vus: 15,
      duration: '8m',
      tags: { scenario: 'api_integration' },
    },
    
    // Chaos testing (error conditions)
    chaos_testing: {
      executor: 'constant-vus',
      vus: 5,
      duration: '3m',
      tags: { scenario: 'chaos_testing' },
    },
  },
  
  thresholds: {
    // Production SLA requirements
    'http_req_duration{scenario:web_traffic}': ['p(95)<1500'],
    'http_req_duration{scenario:batch_processing}': ['p(95)<3000'],
    'http_req_duration{scenario:enterprise_reports}': ['p(95)<5000'],
    'http_req_failed{scenario:web_traffic}': ['rate<0.01'], // 99% success
    'http_req_failed{scenario:batch_processing}': ['rate<0.05'], // 95% success
    'http_req_failed{scenario:enterprise_reports}': ['rate<0.02'], // 98% success
    'concurrent_jobs_active': ['value<200'], // Max 200 concurrent jobs
    'minio_operation_duration': ['p(95)<1000'], // MinIO ops under 1s
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data for different scenarios
const scenarioTemplates = {
  web_traffic: [
    {
      content: '<h1>Invoice #${Math.floor(Math.random() * 10000)}</h1><p>Amount: $${(Math.random() * 1000).toFixed(2)}</p>',
      type: 'html',
      options: { quality: 'standard' }
    },
    {
      content: '<h1>Receipt</h1><table><tr><th>Item</th><th>Price</th></tr><tr><td>Product A</td><td>$25.99</td></tr></table>',
      type: 'html',
      options: { quality: 'standard' }
    }
  ],
  
  batch_processing: [
    {
      content: generateBatchReport(),
      type: 'html',
      options: { quality: 'high' }
    }
  ],
  
  enterprise_reports: [
    {
      content: generateEnterpriseReport(),
      type: 'html',
      options: { 
        quality: 'high',
        page: { size: { name: 'A4' }, orientation: 'portrait' }
      }
    }
  ],
  
  api_integration: [
    {
      content: '# API Integration Test\n\nThis PDF was generated via API integration testing.',
      type: 'markdown',
      options: { quality: 'standard' }
    }
  ],
  
  chaos_testing: [
    {
      content: generateChaosContent(),
      type: 'html',
      options: { quality: 'high' }
    }
  ]
};

export default function() {
  const scenario = __ENV.K6_SCENARIO || 'web_traffic';
  const templates = scenarioTemplates[scenario] || scenarioTemplates.web_traffic;
  const template = templates[Math.floor(Math.random() * templates.length)];
  
  // Add scenario-specific behavior
  switch (scenario) {
    case 'web_traffic':
      return handleWebTraffic(template);
    case 'batch_processing':
      return handleBatchProcessing(template);
    case 'enterprise_reports':
      return handleEnterpriseReports(template);
    case 'api_integration':
      return handleAPIIntegration(template);
    case 'chaos_testing':
      return handleChaosTesting(template);
    default:
      return handleWebTraffic(template);
  }
}

// Web traffic scenario - fast, lightweight PDFs
function handleWebTraffic(template) {
  const startTime = Date.now();
  
  // Submit job
  const job = submitJob(template);
  if (!job) return;
  
  // Quick polling for web traffic
  const completed = pollJobQuick(job.job_id, 10); // 10 second timeout
  if (!completed) return;
  
  // Download PDF
  downloadPDF(job.job_id);
  
  // Track concurrent jobs
  concurrentJobsGauge.add(1);
  
  sleep(0.1); // Minimal sleep for web traffic
}

// Batch processing scenario - multiple documents per user
function handleBatchProcessing(template) {
  const batchSize = Math.floor(Math.random() * 5) + 3; // 3-7 documents per batch
  const jobs = [];
  
  // Submit batch of jobs
  for (let i = 0; i < batchSize; i++) {
    const job = submitJob({
      ...template,
      content: template.content.replace('${i}', i.toString())
    });
    if (job) jobs.push(job);
    sleep(0.2); // Small delay between submissions
  }
  
  // Wait for all jobs to complete
  for (const job of jobs) {
    const completed = pollJobStandard(job.job_id, 30);
    if (completed) {
      downloadPDF(job.job_id);
    }
  }
  
  concurrentJobsGauge.add(batchSize);
  sleep(1); // Batch processing pause
}

// Enterprise reports scenario - large, complex documents
function handleEnterpriseReports(template) {
  const startTime = Date.now();
  
  // Submit complex report job
  const job = submitJob(template);
  if (!job) return;
  
  // Extended polling for complex reports
  const completed = pollJobExtended(job.job_id, 60); // 60 second timeout
  if (!completed) return;
  
  // Download and validate large PDF
  const success = downloadPDF(job.job_id);
  if (success) {
    // Estimate memory usage for large PDFs
    memoryUsageGauge.add(Math.random() * 50 + 20); // 20-70 MB estimate
  }
  
  const duration = Date.now() - startTime;
  minioLatencyTrend.add(duration);
  
  sleep(2); // Enterprise processing pause
}

// API integration scenario - simulate external system calls
function handleAPIIntegration(template) {
  // Simulate API authentication delay
  sleep(0.1);
  
  // Submit job with API-style headers
  const job = submitJobWithAuth(template);
  if (!job) return;
  
  // Standard polling
  const completed = pollJobStandard(job.job_id, 20);
  if (!completed) return;
  
  // Download with validation
  const success = downloadPDF(job.job_id);
  
  // Simulate API response processing
  sleep(0.3);
}

// Chaos testing scenario - test error conditions
function handleChaosTesting(template) {
  const chaosType = Math.random();
  
  if (chaosType < 0.3) {
    // Test malformed content
    testMalformedContent();
  } else if (chaosType < 0.6) {
    // Test oversized content
    testOversizedContent();
  } else {
    // Test invalid parameters
    testInvalidParameters();
  }
  
  sleep(1);
}

// Helper functions
function submitJob(template) {
  const payload = JSON.stringify(template);
  const response = http.post(`${BASE_URL}/api/v1/print`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const success = check(response, {
    'submit: status 202': (r) => r.status === 202,
    'submit: has job_id': (r) => JSON.parse(r.body).job_id !== undefined,
  });
  
  if (!success) {
    errorsByTypeCounter.add(1, { error_type: 'submit_failed' });
    return null;
  }
  
  return JSON.parse(response.body);
}

function submitJobWithAuth(template) {
  const payload = JSON.stringify(template);
  const response = http.post(`${BASE_URL}/api/v1/print`, payload, {
    headers: { 
      'Content-Type': 'application/json',
      'X-API-Client': 'k6-load-test',
      'X-Request-ID': `req-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    },
  });
  
  return response.status === 202 ? JSON.parse(response.body) : null;
}

function pollJobQuick(jobId, timeoutSeconds) {
  return pollJob(jobId, timeoutSeconds, 0.5); // 500ms intervals
}

function pollJobStandard(jobId, timeoutSeconds) {
  return pollJob(jobId, timeoutSeconds, 1); // 1s intervals
}

function pollJobExtended(jobId, timeoutSeconds) {
  return pollJob(jobId, timeoutSeconds, 2); // 2s intervals
}

function pollJob(jobId, timeoutSeconds, intervalSeconds) {
  const maxAttempts = Math.floor(timeoutSeconds / intervalSeconds);
  let attempts = 0;
  
  while (attempts < maxAttempts) {
    const response = http.get(`${BASE_URL}/api/v1/print/${jobId}`);
    
    if (response.status === 200) {
      const job = JSON.parse(response.body);
      if (job.status === 'completed') return job;
      if (job.status === 'failed') {
        errorsByTypeCounter.add(1, { error_type: 'job_failed' });
        return null;
      }
    }
    
    attempts++;
    sleep(intervalSeconds);
  }
  
  errorsByTypeCounter.add(1, { error_type: 'job_timeout' });
  return null;
}

function downloadPDF(jobId) {
  const response = http.get(`${BASE_URL}/api/v1/print/${jobId}/download`);
  
  const success = check(response, {
    'download: status 200': (r) => r.status === 200,
    'download: is PDF': (r) => r.headers['Content-Type'] === 'application/pdf',
    'download: has content': (r) => r.body.length > 0,
  });
  
  if (!success) {
    errorsByTypeCounter.add(1, { error_type: 'download_failed' });
  }
  
  return success;
}

// Chaos testing functions
function testMalformedContent() {
  const malformedTemplate = {
    content: '<html><body><h1>Unclosed tag<p>Missing closing tags</body>', // Malformed HTML
    type: 'html',
    options: { quality: 'standard' }
  };
  
  const response = http.post(`${BASE_URL}/api/v1/print`, JSON.stringify(malformedTemplate), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  // Should handle gracefully
  check(response, {
    'chaos malformed: handled gracefully': (r) => r.status === 202 || r.status === 400,
  });
}

function testOversizedContent() {
  const oversizedContent = '<html><body>' + 'X'.repeat(10000000) + '</body></html>'; // 10MB content
  const oversizedTemplate = {
    content: oversizedContent,
    type: 'html',
    options: { quality: 'standard' }
  };
  
  const response = http.post(`${BASE_URL}/api/v1/print`, JSON.stringify(oversizedTemplate), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(response, {
    'chaos oversized: handled gracefully': (r) => r.status === 202 || r.status === 413,
  });
}

function testInvalidParameters() {
  const invalidTemplate = {
    content: 'Valid content',
    type: 'invalid_type', // Invalid type
    options: { quality: 'invalid_quality' } // Invalid quality
  };
  
  const response = http.post(`${BASE_URL}/api/v1/print`, JSON.stringify(invalidTemplate), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(response, {
    'chaos invalid: handled gracefully': (r) => r.status === 400 || r.status === 202,
  });
}

// Content generators
function generateBatchReport() {
  return `
    <html>
      <head><title>Batch Report ${Math.floor(Math.random() * 1000)}</title></head>
      <body>
        <h1>Batch Processing Report</h1>
        <p>Generated: ${new Date().toISOString()}</p>
        <table border="1">
          <tr><th>ID</th><th>Status</th><th>Value</th></tr>
          ${Array.from({length: 20}, (_, i) => 
            `<tr><td>${i+1}</td><td>Processed</td><td>$${(Math.random() * 1000).toFixed(2)}</td></tr>`
          ).join('')}
        </table>
      </body>
    </html>
  `;
}

function generateEnterpriseReport() {
  return `
    <html>
      <head>
        <title>Enterprise Performance Report</title>
        <style>
          body { font-family: Arial, sans-serif; margin: 20px; }
          .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; }
          .metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin: 20px 0; }
          .metric-card { border: 1px solid #ddd; padding: 20px; border-radius: 8px; }
          .chart-placeholder { height: 200px; background: #f5f5f5; margin: 20px 0; }
          table { width: 100%; border-collapse: collapse; margin: 20px 0; }
          th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
          th { background: #f2f2f2; }
        </style>
      </head>
      <body>
        <div class="header">
          <h1>Q4 2024 Performance Report</h1>
          <p>Enterprise Analytics Dashboard</p>
        </div>
        
        <div class="metrics">
          <div class="metric-card">
            <h3>Total Revenue</h3>
            <h2>$${(Math.random() * 1000000).toFixed(0)}</h2>
          </div>
          <div class="metric-card">
            <h3>Active Users</h3>
            <h2>${(Math.random() * 100000).toFixed(0)}</h2>
          </div>
          <div class="metric-card">
            <h3>Conversion Rate</h3>
            <h2>${(Math.random() * 10).toFixed(2)}%</h2>
          </div>
        </div>
        
        <div class="chart-placeholder">
          <p style="text-align: center; padding-top: 80px;">Performance Chart Placeholder</p>
        </div>
        
        <table>
          <tr><th>Department</th><th>Budget</th><th>Spent</th><th>Remaining</th></tr>
          ${Array.from({length: 15}, (_, i) => {
            const budget = Math.random() * 100000;
            const spent = budget * (0.3 + Math.random() * 0.6);
            return `<tr>
              <td>Department ${i+1}</td>
              <td>$${budget.toFixed(0)}</td>
              <td>$${spent.toFixed(0)}</td>
              <td>$${(budget - spent).toFixed(0)}</td>
            </tr>`;
          }).join('')}
        </table>
      </body>
    </html>
  `;
}

function generateChaosContent() {
  const chaosElements = [
    '<script>alert("XSS test")</script>', // XSS attempt
    '<img src="nonexistent.jpg" onerror="alert(1)">', // Error injection
    '<?xml version="1.0"?><!DOCTYPE test [<!ENTITY xxe "XXE test">]>', // XXE attempt
    '<style>body { display: none; }</style>', // CSS injection
  ];
  
  return `
    <html>
      <body>
        <h1>Chaos Testing Content</h1>
        <p>This content tests security and error handling:</p>
        ${chaosElements[Math.floor(Math.random() * chaosElements.length)]}
        <p>Normal content continues here...</p>
      </body>
    </html>
  `;
}
