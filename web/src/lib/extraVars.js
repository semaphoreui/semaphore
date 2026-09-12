// Conversions between the table editor rows of an environment's extra
// variables and their JSON representation. Pure functions, used by
// EnvironmentForm.vue.

// isPlainObject is true only for a plain name -> value map: the shape
// extra variables must have at the root.
export function isPlainObject(value) {
  if (value === null || Array.isArray(value) || typeof value !== 'object') {
    return false;
  }
  const proto = Object.getPrototypeOf(value);
  return proto === Object.prototype || proto === null;
}

// isJsonSafeValue recursively rejects NaN/Infinity/-Infinity anywhere in
// value: JSON.stringify silently turns these into null instead of
// throwing, so they'd otherwise pass every other check and still corrupt
// the saved data. seen avoids infinite recursion on a circular
// reference; that case is left to throw naturally so the caller reports it.
export function isJsonSafeValue(value, seen = new Set()) {
  if (typeof value === 'number') {
    return Number.isFinite(value);
  }
  if (Array.isArray(value)) {
    if (seen.has(value)) {
      return true;
    }
    seen.add(value);
    return value.every((v) => isJsonSafeValue(v, seen));
  }
  if (isPlainObject(value)) {
    if (seen.has(value)) {
      return true;
    }
    seen.add(value);
    return Object.values(value).every((v) => isJsonSafeValue(v, seen));
  }
  return true;
}

// inferVarType maps a parsed JSON value to one of the editor's variable types
// so that a value stored as a list/dict is shown with the right type in the
// table editor. Numbers get the number type; other scalars are edited as strings.
export function inferVarType(value) {
  if (Array.isArray(value)) {
    return 'list';
  }
  if (value !== null && typeof value === 'object') {
    return 'dict';
  }
  if (typeof value === 'number') {
    return 'number';
  }
  return 'string';
}

// rowToVarValue converts a table row back to its typed JSON value. It throws a
// descriptive error when the input is invalid.
export function rowToVarValue(row) {
  switch (row.type) {
    case 'number': {
      const parsed = Number(row.value);
      if (row.value === '' || Number.isNaN(parsed)) {
        throw new Error(`Variable "${row.name}" must be a number, e.g. 42`);
      }
      return parsed;
    }
    case 'list': {
      let parsed;
      try {
        parsed = JSON.parse(row.value);
      } catch (e) {
        throw new Error(`Variable "${row.name}" is not a valid list, e.g. ["a", "b"]`);
      }
      if (!Array.isArray(parsed)) {
        throw new Error(`Variable "${row.name}" must be a list, e.g. ["a", "b"]`);
      }
      return parsed;
    }
    case 'dict': {
      let parsed;
      try {
        parsed = JSON.parse(row.value);
      } catch (e) {
        throw new Error(`Variable "${row.name}" is not a valid dict, e.g. {"key": "value"}`);
      }
      if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
        throw new Error(`Variable "${row.name}" must be a dict, e.g. {"key": "value"}`);
      }
      return parsed;
    }
    default:
      // Scalars are passed through as-is: an untouched number/boolean keeps its
      // original JSON type, while typed input stays a string.
      return row.value;
  }
}

export function extraVarsToObject(rows) {
  return (rows || []).reduce(
    (prev, curr) => ({
      ...prev,
      [curr.name]: rowToVarValue(curr),
    }),
    {},
  );
}

// extraVarsToObjectLenient is like extraVarsToObject but never throws: a row
// whose typed value is still invalid keeps its raw text. Used when toggling to
// the JSON view so editing modes never lose data mid-typing.
export function extraVarsToObjectLenient(rows) {
  return (rows || []).reduce((prev, curr) => {
    let value;
    try {
      value = rowToVarValue(curr);
    } catch (e) {
      value = curr.value;
    }
    return { ...prev, [curr.name]: value };
  }, {});
}

export function objectToExtraVars(obj) {
  return Object.keys(obj).map((name) => {
    const value = obj[name];
    const type = inferVarType(value);
    return {
      name,
      type,
      value: type === 'string' ? value : JSON.stringify(value),
    };
  });
}
