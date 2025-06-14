/* eslint-disable no-underscore-dangle,no-param-reassign */
import dayjs from 'dayjs';
import isoWeek from 'dayjs/plugin/isoWeek'; // for isoWeekday functionality
import customParseFormat from 'dayjs/plugin/customParseFormat'; // for parsing with format
import { _adapters } from 'chart.js';

// Register required Day.js plugins
dayjs.extend(isoWeek);
dayjs.extend(customParseFormat);

const FORMATS = {
  datetime: 'MMM D, YYYY, h:mm:ss a',
  millisecond: 'h:mm:ss.SSS a',
  second: 'h:mm:ss a',
  minute: 'h:mm a',
  hour: 'hA',
  day: 'MMM D',
  week: 'll',
  month: 'MMM YYYY',
  quarter: '[Q]Q - YYYY',
  year: 'YYYY',
};

_adapters._date.override(typeof dayjs === 'function' ? {
  _id: 'dayjs', // DEBUG ONLY

  formats() {
    return FORMATS;
  },

  parse(value, format) {
    if (typeof value === 'string' && typeof format === 'string') {
      value = dayjs(value, format);
    } else {
      value = dayjs(value);
    }
    return value.isValid() ? value.valueOf() : null;
  },

  format(time, format) {
    return dayjs(time).format(format);
  },

  add(time, amount, unit) {
    return dayjs(time).add(amount, unit).valueOf();
  },

  diff(max, min, unit) {
    return dayjs(max).diff(dayjs(min), unit);
  },

  startOf(time, unit, weekday) {
    time = dayjs(time);
    if (unit === 'isoWeek') {
      // Day.js counts from 1 (Monday) to 7 (Sunday)
      weekday = Math.trunc(Math.min(Math.max(1, weekday || 1), 7));
      return time.isoWeekday(weekday).startOf('day').valueOf();
    }
    return time.startOf(unit).valueOf();
  },

  endOf(time, unit) {
    return dayjs(time).endOf(unit).valueOf();
  },
} : {});
