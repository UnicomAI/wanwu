import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Helper to read and eval file content to get the object
// We can't directly import because of potential environment issues with Vue/Webpack specific syntax if any
// But these files look like pure JS objects.
// However, they use 'export default', so we can strip that and eval.

function loadLangFile(filePath) {
    const content = fs.readFileSync(filePath, 'utf8');
    // Remove "export default" and any leading comments/whitespace
    const objectContent = content.replace(/export default\s*/, '').replace(/\/\/.*/g, '');
    // We need to make it a valid expression to eval. 
    // Note: eval is dangerous but we trust these files.
    // We might need to handle 'require' if used.
    const require = () => 'DUMMY_REQUIRE'; // Mock require
    try {
        return eval('(' + objectContent + ')');
    } catch (e) {
        console.error(`Error parsing ${filePath}:`, e);
        return {};
    }
}

const zhPath = path.resolve(__dirname, '../web/src/lang/zh.js');
const enPath = path.resolve(__dirname, '../web/src/lang/en.js');

const zh = loadLangFile(zhPath);
const en = loadLangFile(enPath);

function getKeys(obj, prefix = '') {
  let keys = [];
  for (const key in obj) {
    const newPrefix = prefix ? `${prefix}.${key}` : key;
    // If it's an object and not an array, recurse
    if (typeof obj[key] === 'object' && obj[key] !== null && !Array.isArray(obj[key])) {
        // Check if it's a leaf node (some keys might be objects but treated as values in some contexts? 
        // usually i18n files are nested objects until string)
        // But wait, if one is string and other is object, that's a mismatch too.
        keys.push(newPrefix); // Add the object key itself too
        keys = keys.concat(getKeys(obj[key], newPrefix));
    } else {
        keys.push(newPrefix);
    }
  }
  return keys;
}

const zhKeys = new Set(getKeys(zh));
const enKeys = new Set(getKeys(en));

const missingInEn = [...zhKeys].filter(k => !enKeys.has(k));
const missingInZh = [...enKeys].filter(k => !zhKeys.has(k));

console.log('Missing in EN:', missingInEn);
console.log('Missing in ZH:', missingInZh);
