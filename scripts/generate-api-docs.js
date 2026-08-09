#!/usr/bin/env node
/**
 * OpenAPI to Markdown Generator
 * 
 * Converts OpenAPI 3.1 spec to formatted Markdown documentation
 * with multi-language code samples for the admin dashboard.
 * 
 * Usage: node scripts/generate-api-docs.js
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

const OPENAPI_FILE = path.join(__dirname, '../docs/openapi/openapi.yaml');
const OUTPUT_DIR = path.join(__dirname, '../docs/openapi/generated');
const OUTPUT_FILE = path.join(OUTPUT_DIR, 'api-reference.md');

/**
 * Load and parse OpenAPI spec
 */
function loadOpenAPISpec() {
    const content = fs.readFileSync(OPENAPI_FILE, 'utf8');
    return yaml.load(content);
}

/**
 * Generate code sample for curl
 */
function generateCurlSample(path, method, operation, servers) {
    const server = servers[0]?.url || 'https://api.clpr.tv';
    const url = `${server}${path}`;
    
    let sample = `curl -X ${method.toUpperCase()} "${url}"`;
    
    // Add auth header if required
    if (operation.security && operation.security.length > 0) {
        sample += ` \\\n  -H "Authorization: Bearer YOUR_TOKEN"`;
    }
    
    // Add content-type header only when a request body is defined
    if (operation.requestBody && ['post', 'put', 'patch'].includes(method.toLowerCase())) {
        sample += ` \\\n  -H "Content-Type: application/json"`;
    }
    
    // Add example request body
    if (operation.requestBody) {
        sample += ` \\\n  -d '{"example": "data"}'`;
    }
    
    return sample;
}

/**
 * Generate code sample for JavaScript/TypeScript
 */
function generateJavaScriptSample(path, method, operation) {
    const hasAuth = operation.security && operation.security.length > 0;
    
    let sample = `// Using fetch API\n`;
    sample += `try {\n`;
    sample += `  const response = await fetch('${path}', {\n`;
    sample += `    method: '${method.toUpperCase()}',\n`;
    
    if (hasAuth || operation.requestBody) {
        sample += `    headers: {\n`;
        if (hasAuth) {
            sample += `      'Authorization': 'Bearer YOUR_TOKEN',\n`;
        }
        if (operation.requestBody) {
            sample += `      'Content-Type': 'application/json'\n`;
        }
        sample += `    }`;
        if (operation.requestBody) {
            sample += `,\n    body: JSON.stringify({\n      // Your request data\n    })`;
        }
        sample += `\n`;
    }
    
    sample += `  });\n`;
    sample += `  \n  if (!response.ok) {\n`;
    sample += `    throw new Error('HTTP error ' + response.status);\n`;
    sample += `  }\n`;
    sample += `  \n  const data = await response.json();\n`;
    sample += `  // Process data\n`;
    sample += `} catch (error) {\n`;
    sample += `  console.error('Error:', error);\n`;
    sample += `}`;
    
    return sample;
}

/**
 * Generate code sample for Python
 */
function generatePythonSample(path, method, operation) {
    const hasAuth = operation.security && operation.security.length > 0;
    
    let sample = `import requests\n\n`;
    
    if (hasAuth) {
        sample += `headers = {'Authorization': 'Bearer YOUR_TOKEN'}\n`;
    }
    
    sample += `try:\n`;
    sample += `    response = requests.${method.toLowerCase()}(\n`;
    sample += `        '${path}'`;
    
    if (hasAuth) {
        sample += `,\n        headers=headers`;
    }
    
    if (operation.requestBody) {
        sample += `,\n        json={}  # Your request data`;
    }
    
    sample += `\n    )\n`;
    sample += `    response.raise_for_status()  # Raise error for bad status\n`;
    sample += `    data = response.json()\n`;
    sample += `    # Process data\n`;
    sample += `except requests.exceptions.RequestException as e:\n`;
    sample += `    print(f"Error: {e}")`;
    
    return sample;
}

/**
 * Generate code sample for Go
 */
function generateGoSample(path, method, operation) {
    const hasAuth = operation.security && operation.security.length > 0;
    const hasBody = operation.requestBody;
    
    let imports = '"net/http"\n    "io"';
    if (hasBody) {
        imports += '\n    "bytes"\n    "encoding/json"';
    }
    
    let sample = `package main\n\nimport (\n    ${imports}\n)\n\n`;
    sample += `func main() {\n`;
    
    // Create request body if needed
    if (hasBody) {
        sample += `    // Create request body\n`;
        sample += `    data := map[string]interface{}{\n`;
        sample += `        // Your request data\n`;
        sample += `    }\n`;
        sample += `    jsonBody, err := json.Marshal(data)\n`;
        sample += `    if err != nil {\n`;
        sample += `        // Handle error\n`;
        sample += `        return\n`;
        sample += `    }\n`;
        sample += `    \n`;
        sample += `    req, err := http.NewRequest("${method.toUpperCase()}", "${path}", bytes.NewBuffer(jsonBody))\n`;
    } else {
        sample += `    req, err := http.NewRequest("${method.toUpperCase()}", "${path}", nil)\n`;
    }
    
    sample += `    if err != nil {\n`;
    sample += `        // Handle error\n`;
    sample += `        return\n`;
    sample += `    }\n`;
    
    if (hasAuth) {
        sample += `    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")\n`;
    }
    
    if (hasBody) {
        sample += `    req.Header.Set("Content-Type", "application/json")\n`;
    }
    
    sample += `    \n    client := &http.Client{}\n`;
    sample += `    resp, err := client.Do(req)\n`;
    sample += `    if err != nil {\n`;
    sample += `        // Handle error\n`;
    sample += `        return\n`;
    sample += `    }\n`;
    sample += `    defer resp.Body.Close()\n`;
    sample += `    body, err := io.ReadAll(resp.Body)\n`;
    sample += `    if err != nil {\n`;
    sample += `        // Handle error\n`;
    sample += `        return\n`;
    sample += `    }\n`;
    sample += `    // Process body\n`;
    sample += `    _ = body\n`;
    sample += `}`;
    
    return sample;
}

/**
 * Generate markdown for an endpoint
 */
function generateEndpointMarkdown(path, method, operation, servers) {
    let md = '';
    
    // Endpoint title
    const operationId = operation.operationId || `${method}_${path}`;
    md += `### ${operation.summary || operationId}\n\n`;
    
    // Method and path badge
    md += `\`${method.toUpperCase()} ${path}\`\n\n`;
    
    // Description
    if (operation.description) {
        md += `${operation.description}\n\n`;
    }
    
    // Tags
    if (operation.tags && operation.tags.length > 0) {
        md += `**Tags:** ${operation.tags.join(', ')}\n\n`;
    }
    
    // Security
    if (operation.security && operation.security.length > 0) {
        md += `🔒 **Authentication Required**\n\n`;
    }
    
    // Parameters
    if (operation.parameters && operation.parameters.length > 0) {
        md += `#### Parameters\n\n`;
        md += `| Name | In | Type | Required | Description |\n`;
        md += `|------|-------|------|----------|-------------|\n`;
        
        operation.parameters.forEach(param => {
            const name = param.name || '';
            const inLocation = param.in || '';
            const type = param.schema?.type || 'string';
            const required = param.required ? '✓' : '';
            const description = (param.description || '').replace(/\n/g, ' ');
            
            md += `| ${name} | ${inLocation} | ${type} | ${required} | ${description} |\n`;
        });
        
        md += `\n`;
    }
    
    // Request Body
    if (operation.requestBody) {
        md += `#### Request Body\n\n`;
        const content = operation.requestBody.content;
        if (content && content['application/json']) {
            md += `Content-Type: \`application/json\`\n\n`;
            if (operation.requestBody.description) {
                md += `${operation.requestBody.description}\n\n`;
            }
        }
    }
    
    // Responses
    if (operation.responses) {
        md += `#### Responses\n\n`;
        
        Object.entries(operation.responses).forEach(([statusCode, response]) => {
            md += `**${statusCode}** - ${response.description || 'Success'}\n\n`;
        });
    }
    
    // Code Samples
    md += `#### Code Examples\n\n`;
    
    // cURL
    md += `##### cURL\n\n`;
    md += `\`\`\`bash\n${generateCurlSample(path, method, operation, servers)}\n\`\`\`\n\n`;
    
    // JavaScript
    md += `##### JavaScript\n\n`;
    md += `\`\`\`javascript\n${generateJavaScriptSample(path, method, operation)}\n\`\`\`\n\n`;
    
    // Python
    md += `##### Python\n\n`;
    md += `\`\`\`python\n${generatePythonSample(path, method, operation)}\n\`\`\`\n\n`;
    
    // Go
    md += `##### Go\n\n`;
    md += `\`\`\`go\n${generateGoSample(path, method, operation)}\n\`\`\`\n\n`;
    
    md += `---\n\n`;
    
    return md;
}

/**
 * Generate complete API reference markdown
 */
function generateAPIReference(spec) {
    let md = '';
    
    // Header
    md += `---\n`;
    md += `title: "API Reference"\n`;
    md += `summary: "Complete API reference for Clipper platform"\n`;
    md += `tags: ["api", "reference", "openapi"]\n`;
    md += `area: "openapi"\n`;
    md += `status: "stable"\n`;
    md += `version: "${spec.info.version}"\n`;
    md += `generated: ${new Date().toISOString()}\n`;
    md += `---\n\n`;
    
    // Title and description
    md += `# ${spec.info.title}\n\n`;
    md += `${spec.info.description}\n\n`;
    md += `**Version:** ${spec.info.version}\n\n`;
    
    // Servers
    if (spec.servers && spec.servers.length > 0) {
        md += `## Base URLs\n\n`;
        spec.servers.forEach(server => {
            md += `- **${server.description || 'API Server'}:** \`${server.url}\`\n`;
        });
        md += `\n`;
    }
    
    // Table of Contents by Tag
    md += `## Table of Contents\n\n`;
	const usedTags = new Set();
	Object.values(spec.paths || {}).forEach(pathItem => {
		Object.values(pathItem || {}).forEach(operation => {
			if (!operation || typeof operation !== 'object') return;
			(operation.tags || ['Uncategorized']).forEach(tag => usedTags.add(tag));
		});
	});
    if (spec.tags && spec.tags.length > 0) {
		spec.tags.filter(tag => usedTags.has(tag.name)).forEach(tag => {
            md += `- [${tag.name}](#${tag.name.toLowerCase().replace(/\s+/g, '-')})\n`;
        });
        md += `\n`;
    }
    
    // Group endpoints by tag
    const endpointsByTag = {};
    
    Object.entries(spec.paths || {}).forEach(([path, pathItem]) => {
        ['get', 'post', 'put', 'patch', 'delete'].forEach(method => {
            const operation = pathItem[method];
            if (operation) {
                const tags = operation.tags || ['Uncategorized'];
                tags.forEach(tag => {
                    if (!endpointsByTag[tag]) {
                        endpointsByTag[tag] = [];
                    }
                    endpointsByTag[tag].push({ path, method, operation });
                });
            }
        });
    });
    
    // Generate sections by tag
    Object.entries(endpointsByTag).forEach(([tag, endpoints]) => {
        md += `## ${tag}\n\n`;
        
        // Find tag description
        const tagInfo = spec.tags?.find(t => t.name === tag);
        if (tagInfo && tagInfo.description) {
            md += `${tagInfo.description}\n\n`;
        }
        
        // Generate endpoint documentation
        endpoints.forEach(({ path, method, operation }) => {
            md += generateEndpointMarkdown(path, method, operation, spec.servers || []);
        });
    });
    
    return md;
}

/**
 * Main function
 */
function main() {
    console.log('🚀 Generating API documentation from OpenAPI spec...');
    
    try {
        // Load OpenAPI spec
        console.log('📖 Loading OpenAPI spec...');
        const spec = loadOpenAPISpec();
        
        // Create output directory
        if (!fs.existsSync(OUTPUT_DIR)) {
            fs.mkdirSync(OUTPUT_DIR, { recursive: true });
        }
        
        // Generate markdown
        console.log('📝 Generating Markdown...');
        // Keep generated documentation compatible with the repository's
        // whitespace gate even when code-sample templates contain spacer lines.
        const markdown = generateAPIReference(spec).replace(/[ \t]+$/gm, '');
        
        // Write to file
        fs.writeFileSync(OUTPUT_FILE, markdown, 'utf8');
        
        console.log(`✅ API documentation generated successfully!`);
        console.log(`📄 Output: ${OUTPUT_FILE}`);
        console.log(`📊 Size: ${(markdown.length / 1024).toFixed(2)} KB`);
        
        // Generate summary
        const endpointCount = Object.keys(spec.paths || {}).reduce((count, path) => {
            return count + Object.keys(spec.paths[path]).filter(k => 
                ['get', 'post', 'put', 'patch', 'delete'].includes(k)
            ).length;
        }, 0);
        
        console.log(`🔢 Documented ${endpointCount} endpoints`);
        console.log(`🏷️  Organized into ${spec.tags?.length || 0} categories`);
        
    } catch (error) {
        console.error('❌ Error generating API documentation:', error);
        process.exit(1);
    }
}

// Run if called directly
if (require.main === module) {
    main();
}

module.exports = { generateAPIReference, loadOpenAPISpec };
