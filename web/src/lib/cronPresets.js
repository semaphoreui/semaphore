// Helpers for the schedule form's "simple" mode, which edits a 5-field cron
// expression (minute hour day-of-month month day-of-week) through preset
// timings. Pure functions, used by ScheduleForm.vue.

import { CronExpression, CronExpressionParser, CronFieldCollection } from 'cron-parser';

const FIELD = '\\S+';
const SET = '[^*]\\S*'; // a field that is not a bare "*"

const WEEKDAY_NAMES = ['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT'];

// One day-of-week item such as 2#2 or TUE#2 (the second Tuesday) or 5L (the last
// Friday). Like most calendars the form offers the first four weeks and the last.
const WEEKDAY_OF_MONTH = /^([0-7]|[A-Z]{3})(?:#([1-4])|L)$/;

export const LAST = 'last';

// Matches schedules.MaxOffsetDays on the server.
export const MAX_OFFSET_DAYS = 31;

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
    case 'monthly_weekday':
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

/**
 * Reports whether the day fields use the Quartz forms L, W or #, which the
 * timings other than "monthly by weekday" cannot show.
 */
export function usesQuartzDays(cronFormat) {
  const fields = (cronFormat || '').trim().split(/\s+/);
  return fields.length === 5 && (/[LW]/i.test(fields[2]) || /[#L]/i.test(fields[4]));
}

function ordinalRank(ordinal) {
  return ordinal === LAST ? Infinity : ordinal;
}

function parseWeekdayOfMonthItem(item) {
  const match = WEEKDAY_OF_MONTH.exec(item);
  if (!match) {
    return null;
  }

  const weekday = /\d/.test(match[1]) ? Number(match[1]) % 7 : WEEKDAY_NAMES.indexOf(match[1]);
  if (weekday < 0) {
    return null;
  }

  return { weekday, ordinal: match[2] ? Number(match[2]) : LAST };
}

/**
 * Reads the expressions the "monthly by weekday" timing produces: any minute
 * and hour, an open day of month and month, and a day of week that lists every
 * combination of some weekdays and ordinals, such as 2#2, 5L or 1#1,1#3.
 * Returns { weekdays, ordinals } with ordinals of 1-4 or LAST, or null.
 */
export function parseWeekdayOfMonth(cronFormat) {
  const fields = (cronFormat || '').trim().split(/\s+/);

  if (fields.length !== 5 || !['*', '?'].includes(fields[2]) || fields[3] !== '*') {
    return null;
  }

  const items = fields[4].toUpperCase().split(',').map(parseWeekdayOfMonthItem);
  if (items.includes(null)) {
    return null;
  }

  const weekdays = [...new Set(items.map((item) => item.weekday))].sort((a, b) => a - b);
  const ordinals = [...new Set(items.map((item) => item.ordinal))]
    .sort((a, b) => ordinalRank(a) - ordinalRank(b));
  const combinations = new Set(items.map((item) => `${item.weekday}/${item.ordinal}`));

  // The form selects weekdays and ordinals separately, so it can only show
  // a day of week that has all their combinations.
  if (combinations.size !== weekdays.length * ordinals.length) {
    return null;
  }

  return { weekdays, ordinals };
}

/**
 * Builds a "monthly by weekday" expression from non-empty weekday and ordinal
 * selections. Empty minute and hour selections become "*", as in buildCronFormat.
 */
export function buildWeekdayOfMonthCronFormat({
  minutes, hours, weekdays, ordinals,
}) {
  const fields = buildCronFormat({ minutes, hours }).split(' ');
  const sortedOrdinals = [...ordinals].sort((a, b) => ordinalRank(a) - ordinalRank(b));

  fields[4] = [...weekdays]
    .sort((a, b) => a - b)
    .flatMap((weekday) => sortedOrdinals.map((ordinal) => (
      ordinal === LAST ? `${weekday}L` : `${weekday}#${ordinal}`
    )))
    .join(',');

  return fields.join(' ');
}

/**
 * Turns user input into a whole number of days within the allowed offset.
 */
export function clampOffsetDays(value) {
  const days = Math.trunc(Number(value));

  if (!Number.isFinite(days)) {
    return 0;
  }

  return Math.min(MAX_OFFSET_DAYS, Math.max(-MAX_OFFSET_DAYS, days));
}
