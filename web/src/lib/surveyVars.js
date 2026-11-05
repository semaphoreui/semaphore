// Survey answers of a task as the task forms (TaskForm.vue, TaskParamsForm.vue)
// seed and keep them. A bool variable is answered with a real boolean: its
// default is stored on the variable as the string "true" or "false", and an
// answer saved as such a string (a task started through the API) is read back
// as a boolean too. Pure functions.

// boolAnswer reads a bool answer or default: true only for true or "true".
export function boolAnswer(value) {
  return value === true || value === 'true';
}

// surveyDefaults returns the defaults of a survey as task environment entries,
// variable name -> value. Every bool variable gets an entry, false without a
// default, so an untouched checkbox still sends an explicit false.
export function surveyDefaults(vars) {
  const res = {};
  (vars || []).forEach((v) => {
    if (v.type === 'bool') {
      res[v.name] = boolAnswer(v.default_value);
    } else if (v.default_value) {
      res[v.name] = v.default_value;
    }
  });
  return res;
}

// normalizeBoolAnswers returns env with every bool answer as a boolean: a
// string "true" becomes true, anything else that is not a boolean false.
export function normalizeBoolAnswers(vars, env) {
  const res = { ...env };
  (vars || []).forEach((v) => {
    if (v.type === 'bool' && typeof res[v.name] !== 'boolean') {
      res[v.name] = boolAnswer(res[v.name]);
    }
  });
  return res;
}
