// Global Vue filters used in templates (e.g. `{{ task.start | formatDate }}`).
// The formatting functions are exported separately so they can be unit-tested
// and reused outside templates.

import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import localizedFormat from 'dayjs/plugin/localizedFormat';
import durationPlugin from 'dayjs/plugin/duration';
import { AnsiUp } from 'ansi_up';

dayjs.extend(relativeTime);
dayjs.extend(localizedFormat);
dayjs.extend(durationPlugin);

const convert = new AnsiUp();
convert.ansi_colors = [
  [
    { rgb: [85, 85, 85], class_name: 'ansi-black' },
    { rgb: [170, 0, 0], class_name: 'ansi-red' },
    { rgb: [0, 170, 0], class_name: 'ansi-green' },
    { rgb: [255, 204, 102], class_name: 'ansi-yellow' },
    { rgb: [33, 150, 243], class_name: 'ansi-blue' },
    { rgb: [170, 0, 170], class_name: 'ansi-magenta' },
    { rgb: [0, 170, 170], class_name: 'ansi-cyan' },
    { rgb: [170, 170, 170], class_name: 'ansi-white' },
  ],
  [
    { rgb: [85, 85, 85], class_name: 'ansi-bright-black' },
    { rgb: [255, 85, 85], class_name: 'ansi-bright-red' },
    { rgb: [85, 255, 85], class_name: 'ansi-bright-green' },
    { rgb: [255, 255, 85], class_name: 'ansi-bright-yellow' },
    { rgb: [85, 85, 255], class_name: 'ansi-bright-blue' },
    { rgb: [255, 85, 255], class_name: 'ansi-bright-magenta' },
    { rgb: [85, 255, 255], class_name: 'ansi-bright-cyan' },
    { rgb: [255, 255, 255], class_name: 'ansi-bright-white' },
  ],
];

const EMPTY = '—';

// formatDate2: localized date without time
export function formatDate2(value) {
  return value ? dayjs(String(value)).format('LL') : EMPTY;
}

// formatDate: “from now” if today, else localized date+time
export function formatDate(value) {
  if (!value) return EMPTY;
  const date = dayjs(value);
  const now = dayjs();

  if (now.isSame(date, 'day')) {
    return `${date.fromNow()} (${date.format('HH:mm')})`;
  }
  return date.format('L HH:mm');
}

// formatTime: localized time with seconds
export function formatTime(value) {
  return value ? dayjs(String(value)).format('LTS') : EMPTY;
}

// formatLog: ANSI → HTML
export function formatLog(value) {
  return value ? convert.ansi_to_html(String(value)) : value;
}

// formatMilliseconds: humanize a duration (number or numeric string) or a
// [start, end] pair; a missing end means "until now".
export function formatMilliseconds(value) {
  if (value == null || value === '') return EMPTY;

  let ms;

  if (typeof value === 'string') {
    ms = parseInt(value, 10);
  } else if (typeof value === 'number') {
    ms = value;
  } else if (Array.isArray(value)) {
    if (value.length !== 2) {
      throw new Error('formatMilliseconds: invalid value format');
    }
    const [startRaw, endRaw] = value;
    if (startRaw == null || startRaw === '') return EMPTY;
    const start = dayjs(startRaw);
    const end = endRaw == null || endRaw === '' ? dayjs() : dayjs(endRaw);
    ms = end.valueOf() - start.valueOf();
  } else {
    throw new Error('formatMilliseconds: unsupported value type');
  }

  return dayjs.duration(ms).humanize();
}

export const filters = {
  formatDate2,
  formatDate,
  formatTime,
  formatLog,
  formatMilliseconds,
};

export default {
  install(Vue) {
    Object.entries(filters).forEach(([name, fn]) => Vue.filter(name, fn));
  },
};
