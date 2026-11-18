import { CORE_SCHEMA, eventsToAst, parseEvents } from 'js-yaml';

// Top-level variable names from the syntax tree, so tags like !vault do not matter.
// Null when the text cannot be read.
export default function variableNames(text) {
  const source = text || '';
  try {
    const [doc] = eventsToAst(parseEvents(source, {}), { source, schema: CORE_SCHEMA });
    const root = doc && doc.contents;
    return root && root.kind === 'mapping' ? root.items.map((item) => item.key.value) : [];
  } catch {
    return null;
  }
}
