#!/usr/bin/env node
/**
 * OpenAPI Changelog Generator
 * 
 * Compares two OpenAPI spec versions and generates a changelog
 * highlighting added, modified, and removed endpoints.
 * 
 * Usage: node scripts/generate-api-changelog.js [old-spec] [new-spec]
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

const OPENAPI_FILE = path.join(__dirname, '../docs/openapi/openapi.yaml');
const OUTPUT_DIR = path.join(__dirname, '../docs/openapi/generated');

/**
 * Load and parse OpenAPI spec
 */
function loadSpec(filePath) {
    const content = fs.readFileSync(filePath, 'utf8');
    return yaml.load(content);
}

/**
 * Extract all endpoints from a spec
 */
function extractEndpoints(spec) {
    const endpoints = [];
    
    Object.entries(spec.paths || {}).forEach(([path, pathItem]) => {
        ['get', 'post', 'put', 'patch', 'delete'].forEach(method => {
            const operation = pathItem[method];
            if (operation) {
                endpoints.push({
                    path,
                    method: method.toUpperCase(),
                    operationId: operation.operationId,
                    summary: operation.summary,
                    tags: operation.tags || [],
                    deprecated: operation.deprecated || false,
                });
            }
        });
    });
    
    return endpoints;
}

/**
 * Create endpoint key for comparison
 */
function endpointKey(endpoint) {
    return `${endpoint.method} ${endpoint.path}`;
}

/**
 * Compare two specs and generate changelog
 */
function compareSpecs(oldSpec, newSpec) {
    const oldEndpoints = extractEndpoints(oldSpec);
    const newEndpoints = extractEndpoints(newSpec);
    
    const oldKeys = new Map(oldEndpoints.map(e => [endpointKey(e), e]));
    const newKeys = new Map(newEndpoints.map(e => [endpointKey(e), e]));
    
    const changes = {
        added: [],
        removed: [],
        modified: [],
        deprecated: [],
    };
    
    // Find added endpoints
    newKeys.forEach((endpoint, key) => {
        if (!oldKeys.has(key)) {
            changes.added.push(endpoint);
        } else {
            // Check for modifications
            const oldEndpoint = oldKeys.get(key);
            if (oldEndpoint.summary !== endpoint.summary) {
                changes.modified.push({
                    ...endpoint,
                    oldSummary: oldEndpoint.summary,
                });
            }
            // Check for newly deprecated
            if (endpoint.deprecated && !oldEndpoint.deprecated) {
                changes.deprecated.push(endpoint);
            }
        }
    });
    
    // Find removed endpoints
    oldKeys.forEach((endpoint, key) => {
        if (!newKeys.has(key)) {
            changes.removed.push(endpoint);
        }
    });
    
    return changes;
}

/**
 * Generate markdown changelog
 */
function generateChangelog(changes, oldVersion, newVersion) {
    let md = '';
    
    // Header
    md += `---\n`;
    md += `title: "API Changelog"\n`;
    md += `summary: "API changes from ${oldVersion} to ${newVersion}"\n`;
    md += `tags: ["api", "changelog", "openapi"]\n`;
    md += `area: "openapi"\n`;
    md += `generated: ${new Date().toISOString()}\n`;
    md += `---\n\n`;
    
    md += `# API Changelog\n\n`;
    md += `## ${oldVersion} → ${newVersion}\n\n`;
    md += `**Generated:** ${new Date().toLocaleString()}\n\n`;
    
    // Summary
    md += `## Summary\n\n`;
    md += `| Change Type | Count |\n`;
    md += `|-------------|-------|\n`;
    md += `| Added | ${changes.added.length} |\n`;
    md += `| Modified | ${changes.modified.length} |\n`;
    md += `| Removed | ${changes.removed.length} |\n`;
    md += `| Deprecated | ${changes.deprecated.length} |\n`;
    md += `| **Total** | **${changes.added.length + changes.modified.length + changes.removed.length + changes.deprecated.length}** |\n\n`;
    
    // Added endpoints
    if (changes.added.length > 0) {
        md += `## ✅ Added Endpoints (${changes.added.length})\n\n`;
        
        // Group by tag
        const byTag = {};
        changes.added.forEach(endpoint => {
            const tag = endpoint.tags[0] || 'Other';
            if (!byTag[tag]) byTag[tag] = [];
            byTag[tag].push(endpoint);
        });
        
        Object.entries(byTag).forEach(([tag, endpoints]) => {
            md += `### ${tag}\n\n`;
            endpoints.forEach(endpoint => {
                md += `- **${endpoint.method} ${endpoint.path}**\n`;
                if (endpoint.summary) {
                    md += `  - ${endpoint.summary}\n`;
                }
            });
            md += `\n`;
        });
    }
    
    // Modified endpoints
    if (changes.modified.length > 0) {
        md += `## 🔄 Modified Endpoints (${changes.modified.length})\n\n`;
        
        changes.modified.forEach(endpoint => {
            md += `- **${endpoint.method} ${endpoint.path}**\n`;
            md += `  - Old: ${endpoint.oldSummary || 'N/A'}\n`;
            md += `  - New: ${endpoint.summary || 'N/A'}\n`;
        });
        md += `\n`;
    }
    
    // Deprecated endpoints
    if (changes.deprecated.length > 0) {
        md += `## ⚠️ Deprecated Endpoints (${changes.deprecated.length})\n\n`;
        
        changes.deprecated.forEach(endpoint => {
            md += `- **${endpoint.method} ${endpoint.path}**\n`;
            if (endpoint.summary) {
                md += `  - ${endpoint.summary}\n`;
            }
        });
        md += `\n`;
    }
    
    // Removed endpoints
    if (changes.removed.length > 0) {
        md += `## ❌ Removed Endpoints (${changes.removed.length})\n\n`;
        
        // Group by tag
        const byTag = {};
        changes.removed.forEach(endpoint => {
            const tag = endpoint.tags[0] || 'Other';
            if (!byTag[tag]) byTag[tag] = [];
            byTag[tag].push(endpoint);
        });
        
        Object.entries(byTag).forEach(([tag, endpoints]) => {
            md += `### ${tag}\n\n`;
            endpoints.forEach(endpoint => {
                md += `- **${endpoint.method} ${endpoint.path}**\n`;
                if (endpoint.summary) {
                    md += `  - ${endpoint.summary}\n`;
                }
            });
            md += `\n`;
        });
    }
    
    // Migration guide
    if (changes.removed.length > 0 || changes.deprecated.length > 0) {
        md += `## 📋 Migration Guide\n\n`;
        
        if (changes.deprecated.length > 0) {
            md += `### Deprecated Endpoints\n\n`;
            md += `The following endpoints are deprecated and will be removed in a future version. Please update your code to use the recommended alternatives:\n\n`;
            changes.deprecated.forEach(endpoint => {
                md += `- \`${endpoint.method} ${endpoint.path}\` - ${endpoint.summary || 'No summary'}\n`;
            });
            md += `\n`;
        }
        
        if (changes.removed.length > 0) {
            md += `### Removed Endpoints\n\n`;
            md += `The following endpoints have been removed. Please review your code and remove any references:\n\n`;
            changes.removed.forEach(endpoint => {
                md += `- \`${endpoint.method} ${endpoint.path}\` - ${endpoint.summary || 'No summary'}\n`;
            });
            md += `\n`;
        }
    }
    
    return md;
}

/**
 * Main function
 */
function main() {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('ℹ️  Usage: node scripts/generate-api-changelog.js [old-spec] [new-spec]');
        console.log('ℹ️  If no arguments provided, will create a baseline changelog from current spec');
        
        try {
            // Generate baseline changelog
            console.log('📖 Loading current OpenAPI spec...');
            const spec = loadSpec(OPENAPI_FILE);
            
            const endpoints = extractEndpoints(spec);
            
            let md = '';
            md += `---\n`;
            md += `title: "API Reference - Baseline"\n`;
            md += `summary: "Baseline API reference for version ${spec.info.version}"\n`;
            md += `tags: ["api", "changelog", "openapi"]\n`;
            md += `area: "openapi"\n`;
            md += `generated: ${new Date().toISOString()}\n`;
            md += `---\n\n`;
            
            md += `# API Baseline - Version ${spec.info.version}\n\n`;
            md += `**Total Endpoints:** ${endpoints.length}\n\n`;
            
            // Group by tag
            const byTag = {};
            endpoints.forEach(endpoint => {
                const tag = endpoint.tags[0] || 'Other';
                if (!byTag[tag]) byTag[tag] = [];
                byTag[tag].push(endpoint);
            });
            
            md += `## Endpoints by Category\n\n`;
            Object.entries(byTag).forEach(([tag, endpoints]) => {
                md += `### ${tag} (${endpoints.length})\n\n`;
                endpoints.forEach(endpoint => {
                    md += `- **${endpoint.method} ${endpoint.path}**`;
                    if (endpoint.summary) {
                        md += ` - ${endpoint.summary}`;
                    }
                    md += `\n`;
                });
                md += `\n`;
            });
            
            // Create output directory
            if (!fs.existsSync(OUTPUT_DIR)) {
                fs.mkdirSync(OUTPUT_DIR, { recursive: true });
            }
            
            const outputFile = path.join(OUTPUT_DIR, 'api-baseline.md');
            fs.writeFileSync(outputFile, `${md.trimEnd()}\n`, 'utf8');
            
            console.log(`✅ Baseline changelog generated!`);
            console.log(`📄 Output: ${outputFile}`);
            console.log(`🔢 Documented ${endpoints.length} endpoints`);
            
            return;
        } catch (error) {
            console.error('❌ Error generating baseline changelog:', error);
            process.exit(1);
        }
    }
    
    // Compare two specs
    const oldSpecFile = args[0];
    const newSpecFile = args[1] || OPENAPI_FILE;
    
    console.log('🚀 Generating API changelog...');
    console.log(`📖 Old spec: ${oldSpecFile}`);
    console.log(`📖 New spec: ${newSpecFile}`);
    
    try {
        const oldSpec = loadSpec(oldSpecFile);
        const newSpec = loadSpec(newSpecFile);
        
        console.log('🔍 Comparing specs...');
        const changes = compareSpecs(oldSpec, newSpec);
        
        console.log('📝 Generating changelog...');
        const markdown = generateChangelog(
            changes,
            oldSpec.info.version,
            newSpec.info.version
        );
        
        // Create output directory
        if (!fs.existsSync(OUTPUT_DIR)) {
            fs.mkdirSync(OUTPUT_DIR, { recursive: true });
        }
        
        const outputFile = path.join(OUTPUT_DIR, 'api-changelog.md');
        fs.writeFileSync(outputFile, markdown, 'utf8');
        
        console.log(`✅ Changelog generated successfully!`);
        console.log(`📄 Output: ${outputFile}`);
        console.log(`\n📊 Summary:`);
        console.log(`  ✅ Added: ${changes.added.length}`);
        console.log(`  🔄 Modified: ${changes.modified.length}`);
        console.log(`  ❌ Removed: ${changes.removed.length}`);
        console.log(`  ⚠️  Deprecated: ${changes.deprecated.length}`);
        
    } catch (error) {
        console.error('❌ Error generating changelog:', error);
        process.exit(1);
    }
}

// Run if called directly
if (require.main === module) {
    main();
}

module.exports = { compareSpecs, generateChangelog };
