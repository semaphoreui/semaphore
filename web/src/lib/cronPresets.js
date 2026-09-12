// Helpers for the schedule form's "simple" mode, which edits a 5-field cron
// expression (minute hour day-of-month month day-of-week) through preset
// timings. Pure functions, used by ScheduleForm.vue.

import { CronExpression, CronExpressionParser, CronFieldCollection } from 'cron-parser';

const FIELD = '\\S+';
const SET = '[^*]\\S*'; // a field that is not a bare "*"

const WEEKLY = new RegExp(`^${FIELD}\\s${FIELD}\\s${FIELD}\\s${FIELD}\\s${SET}$`);
const YEARLY = new RegExp(`^${FIELD}\\s${FIELD}\\s${FIELD}\\s${SET}\\s${FIELD}$`);
const MONTHLY = new RegExp(`^${FIELD}\\s${FIELD}\\s${SET}\\s${FIELD}\\s${FIELD}$`);
const DAILY = new RegExp(`^${FIELD}\\s${SET}\\s${FIELD}\\s${FIELD}\\s${FIELD}$`);
const HOURLY = new RegExp(`^${SET}\\s${FIELD}\\s${FIELD}\\s${FIELD}\\s${FIELD}$`);

export function isWeekly(s) {
  return WEEKLY.test(s);
}

export function isYearly(s) {
  return YEARLY.test(s);
}

export function isMonthly(s) {
  return MONTHLY.test(s);
}

export function isDaily(s) {
  return DAILY.test(s);
}

export function isHourly(s) {
  return HOURLY.test(s);
}

/**
 * Drops the selections that a timing preset does not use, so that e.g.
 * switching from "monthly" to "hourly" does not keep a stale day-of-month.
 * Returns a new object; the input is not modified.
 */
export function pruneSelectionsForTiming(timing, selections) {
  const result = {
    months: [...(selections.months || [])],
    weekdays: [...(selections.weekdays || [])],
    days: [...(selections.days || [])],
    hours: [...(selections.hours || [])],
    minutes: [...(selections.minutes || [])],
  };

  switch (timing) {
    case 'hourly':
      result.months = [];
      result.weekdays = [];
      result.days = [];
      result.hours = [];
      break;
    case 'daily':
      result.days = [];
      result.months = [];
      result.weekdays = [];
      break;
    case 'monthly':
      result.months = [];
      result.weekdays = [];
      break;
    case 'weekly':
      result.months = [];
      result.days = [];
      break;
    default:
      break;
  }

  return result;
}

/**
 * Builds a cron expression from the selected values. Empty selections
 * become "*".
 */
export function buildCronFormat(selections) {
  const fields = {};

  if ((selections.months || []).length > 0) {
    fields.month = selections.months;
  }
  if ((selections.weekdays || []).length > 0) {
    fields.dayOfWeek = selections.weekdays;
  }
  if ((selections.days || []).length > 0) {
    fields.dayOfMonth = selections.days;
  }
  if ((selections.hours || []).length > 0) {
    fields.hour = selections.hours;
  }
  if ((selections.minutes || []).length > 0) {
    fields.minute = selections.minutes;
  }

  const origFields = CronExpressionParser.parse('* * * * *').fields;
  const modFields = CronFieldCollection.from(origFields, fields);
  return CronExpression.fieldsToExpression(modFields).stringify();
}
