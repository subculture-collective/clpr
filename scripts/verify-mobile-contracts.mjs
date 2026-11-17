import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import Ajv2020 from 'ajv/dist/2020.js';
import addFormats from 'ajv-formats';
import yaml from 'js-yaml';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const directory = path.join(root, 'docs/openapi/mobile');

export function loadContracts() {
  return {
    document: yaml.load(fs.readFileSync(path.join(root, 'docs/openapi/openapi.yaml'), 'utf8')),
    manifest: JSON.parse(fs.readFileSync(path.join(directory, 'manifest.json'), 'utf8')),
  };
}

function resolve(document, value) {
  if (!value?.$ref) return value;
  if (!value.$ref.startsWith('#/')) throw new Error(`Unsupported external response reference: ${value.$ref}`);
  const target = value.$ref.slice(2).split('/').reduce((node, key) => node?.[key.replaceAll('~1', '/').replaceAll('~0', '~')], document);
  if (!target) throw new Error(`Missing reference: ${value.$ref}`);
  return resolve(document, target);
}

export function responseValidator(document, fixture) {
  const operation = document.paths[fixture.path]?.[fixture.method.toLowerCase()];
  if (!operation) throw new Error(`Missing operation: ${fixture.method} ${fixture.path}`);
  const response = resolve(document, operation.responses[String(fixture.status)]);
  const schema = response?.content?.['application/json']?.schema;
  if (!schema) throw new Error(`Missing JSON response schema: ${fixture.path} ${fixture.status}`);
  const ajv = new Ajv2020({ strict: false, allErrors: true, formats: { float: true, double: true, int32: true, int64: true } });
  addFormats(ajv);
  return ajv.compile({ ...schema, components: document.components });
}

export function checkContracts() {
  const { document, manifest } = loadContracts();
  if (!Array.isArray(manifest) || manifest.length === 0) throw new Error('Native fixture manifest is empty');
  const registered = JSON.parse(fs.readFileSync(path.join(root, 'docs/openapi/route-contract-manifest.json'), 'utf8'));
  const names = new Set();
  for (const fixture of manifest) {
    if (!/^[a-z0-9-]+\.json$/.test(fixture.file) || names.has(fixture.file)) throw new Error(`Invalid or duplicate fixture filename: ${fixture.file}`);
    names.add(fixture.file);
    if (!['handler-with-owned-stub', 'constructed-from-wire-types', 'constructed-negative-case'].includes(fixture.provenance)) throw new Error(`Missing provenance: ${fixture.file}`);
    if (!fixture.sources?.length || fixture.sources.some(source => !fs.existsSync(path.join(root, source.split('#')[0])))) throw new Error(`Missing source boundary: ${fixture.file}`);
    if (!registered.some(route => route.method === fixture.method && route.path === fixture.path)) throw new Error(`Unregistered route: ${fixture.file}`);
    const body = JSON.parse(fs.readFileSync(path.join(directory, fixture.file), 'utf8'));
    const validate = responseValidator(document, fixture);
    if (!validate(body)) throw new Error(`${fixture.file}: ${JSON.stringify(validate.errors)}`);
  }
  const files = fs.readdirSync(directory).filter(name => name.endsWith('.json') && name !== 'manifest.json');
  if (files.length !== names.size || files.some(name => !names.has(name))) throw new Error('Unowned or missing native fixture');
  return manifest.length;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  console.log(`Native API contracts: ${checkContracts()} fixtures match registered routes and current OpenAPI response schemas.`);
}
