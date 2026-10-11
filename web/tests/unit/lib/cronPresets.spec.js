import { expect } from 'chai';
import {
  isWeekly,
  isYearly,
  isMonthly,
  isDaily,
  isHourly,
  pruneSelectionsForTiming,
  buildCronFormat,
  parseWeekdayOfMonth,
  buildWeekdayOfMonthCronFormat,
  clampOffsetDays,
  usesQuartzDays,
  LAST,
  MAX_OFFSET_DAYS,
} from '@/lib/cronPresets';

describe('lib/cronPresets', () => {
  describe('timing predicates', () => {
    const tests = [
      {
        cron: '* * * * *', hourly: false, daily: false, monthly: false, yearly: false, weekly: false,
      },
      {
        cron: '30 * * * *', hourly: true, daily: false, monthly: false, yearly: false, weekly: false,
      },
      {
        cron: '0 9 * * *', hourly: true, daily: true, monthly: false, yearly: false, weekly: false,
      },
      {
        cron: '0 9 1 * *', hourly: true, daily: true, monthly: true, yearly: false, weekly: false,
      },
      {
        cron: '0 9 1 6 *', hourly: true, daily: true, monthly: true, yearly: true, weekly: false,
      },
      {
        cron: '0 9 * * 1', hourly: true, daily: true, monthly: false, yearly: false, weekly: true,
      },
      {
        cron: '*/5 * * * *', hourly: false, daily: false, monthly: false, yearly: false, weekly: false,
      },
      {
        cron: '0,30 8-18 * * 1-5', hourly: true, daily: true, monthly: false, yearly: false, weekly: true,
      },
      {
        cron: '@hourly', hourly: false, daily: false, monthly: false, yearly: false, weekly: false,
      },
      {
        cron: '', hourly: false, daily: false, monthly: false, yearly: false, weekly: false,
      },
    ];

    tests.forEach((tt) => {
      it(`classifies "${tt.cron}"`, () => {
        expect(isHourly(tt.cron), 'hourly').to.equal(tt.hourly);
        expect(isDaily(tt.cron), 'daily').to.equal(tt.daily);
        expect(isMonthly(tt.cron), 'monthly').to.equal(tt.monthly);
        expect(isYearly(tt.cron), 'yearly').to.equal(tt.yearly);
        expect(isWeekly(tt.cron), 'weekly').to.equal(tt.weekly);
      });
    });
  });

  describe('pruneSelectionsForTiming', () => {
    const full = {
      months: [6], weekdays: [1], days: [15], hours: [9], minutes: [30],
    };

    const tests = [
      {
        timing: 'hourly',
        expected: {
          months: [], weekdays: [], days: [], hours: [], minutes: [30],
        },
      },
      {
        timing: 'daily',
        expected: {
          months: [], weekdays: [], days: [], hours: [9], minutes: [30],
        },
      },
      {
        timing: 'weekly',
        expected: {
          months: [], weekdays: [1], days: [], hours: [9], minutes: [30],
        },
      },
      {
        timing: 'monthly_weekday',
        expected: {
          months: [], weekdays: [1], days: [], hours: [9], minutes: [30],
        },
      },
      {
        timing: 'monthly',
        expected: {
          months: [], weekdays: [], days: [15], hours: [9], minutes: [30],
        },
      },
      {
        timing: 'yearly',
        expected: {
          months: [6], weekdays: [1], days: [15], hours: [9], minutes: [30],
        },
      },
    ];

    tests.forEach(({ timing, expected }) => {
      it(`keeps only the fields used by ${timing}`, () => {
        expect(pruneSelectionsForTiming(timing, full)).to.deep.equal(expected);
      });
    });

    it('does not modify the input', () => {
      const input = { ...full, months: [6] };
      pruneSelectionsForTiming('hourly', input);
      expect(input.months).to.deep.equal([6]);
    });

    it('treats missing arrays as empty', () => {
      expect(pruneSelectionsForTiming('yearly', {})).to.deep.equal({
        months: [], weekdays: [], days: [], hours: [], minutes: [],
      });
    });
  });

  describe('buildCronFormat', () => {
    const tests = [
      { name: 'everything empty', selections: {}, expected: '* * * * *' },
      { name: 'hourly at :30', selections: { minutes: [30] }, expected: '30 * * * *' },
      { name: 'daily at 09:00', selections: { minutes: [0], hours: [9] }, expected: '0 9 * * *' },
      {
        name: 'weekly on Mon and Fri',
        selections: { minutes: [0], hours: [9], weekdays: [1, 5] },
        expected: '0 9 * * 1,5',
      },
      {
        name: 'monthly on the 1st and 15th',
        selections: { minutes: [0], hours: [9], days: [1, 15] },
        expected: '0 9 1,15 * *',
      },
      {
        name: 'yearly in June',
        selections: {
          minutes: [0], hours: [9], days: [1], months: [6],
        },
        expected: '0 9 1 6 *',
      },
      {
        name: 'multiple hours and minutes',
        selections: { minutes: [0, 30], hours: [8, 12, 18] },
        expected: '0,30 8,12,18 * * *',
      },
    ];

    tests.forEach(({ name, selections, expected }) => {
      it(name, () => {
        expect(buildCronFormat(selections)).to.equal(expected);
      });
    });

    it('produces an expression that the timing predicates recognise', () => {
      const cron = buildCronFormat({ minutes: [0], hours: [9], weekdays: [1] });
      expect(isWeekly(cron)).to.equal(true);
      expect(isMonthly(cron)).to.equal(false);
    });
  });

  describe('usesQuartzDays', () => {
    const tests = [
      { cron: '0 3 L * *', expected: true },
      { cron: '0 3 15w * *', expected: true },
      { cron: '0 3 * * 2#2', expected: true },
      { cron: '0 3 * * 5L', expected: true },
      { cron: '0 3 * JUL WED', expected: false },
      { cron: '0 3 1-7 * 1', expected: false },
      { cron: '@monthly', expected: false },
    ];

    tests.forEach(({ cron, expected }) => {
      it(`checks "${cron}"`, () => {
        expect(usesQuartzDays(cron)).to.equal(expected);
      });
    });
  });

  describe('parseWeekdayOfMonth', () => {
    const tests = [
      { cron: '0 3 * * 2#2', expected: { weekdays: [2], ordinals: [2] } },
      { cron: '30 4 * * 5L', expected: { weekdays: [5], ordinals: [LAST] } },
      { cron: '0 3 ? * tue#2', expected: { weekdays: [2], ordinals: [2] } },
      { cron: '0 3,15 * * 7#1', expected: { weekdays: [0], ordinals: [1] } },
      { cron: '0 3 * * 1L,1#1,1#3', expected: { weekdays: [1], ordinals: [1, 3, LAST] } },
      { cron: '0 3 * * 2#2,4#2,2#4,4#4', expected: { weekdays: [2, 4], ordinals: [2, 4] } },
      { cron: '0 3 * * 2#2,4#3', expected: null },
      { cron: '0 3 * * 2', expected: null },
      { cron: '0 3 1 * 2#2', expected: null },
      { cron: '0 3 * 6 2#2', expected: null },
      { cron: '0 3 * * 2#5', expected: null },
      { cron: '0 3 * * XYZ#2', expected: null },
      { cron: '0 3 L * *', expected: null },
      { cron: '0 3 * * 2#2 2027', expected: null },
      { cron: '', expected: null },
    ];

    tests.forEach(({ cron, expected }) => {
      it(`reads "${cron}"`, () => {
        expect(parseWeekdayOfMonth(cron)).to.deep.equal(expected);
      });
    });
  });

  describe('buildWeekdayOfMonthCronFormat', () => {
    const tests = [
      {
        name: 'second Tuesday at 03:00',
        selections: {
          minutes: [0], hours: [3], weekdays: [2], ordinals: [2],
        },
        expected: '0 3 * * 2#2',
      },
      {
        name: 'last Friday twice a day',
        selections: {
          minutes: [30], hours: [4, 16], weekdays: [5], ordinals: [LAST],
        },
        expected: '30 4,16 * * 5L',
      },
      {
        name: 'first and third Monday and Thursday',
        selections: {
          minutes: [0], hours: [3], weekdays: [1, 4], ordinals: [1, 3],
        },
        expected: '0 3 * * 1#1,1#3,4#1,4#3',
      },
      {
        name: 'no time selected',
        selections: { weekdays: [1], ordinals: [1] },
        expected: '* * * * 1#1',
      },
      {
        name: 'selections in click order',
        selections: {
          minutes: [0], hours: [3], weekdays: [5, 1], ordinals: [LAST, 3, 1],
        },
        expected: '0 3 * * 1#1,1#3,1L,5#1,5#3,5L',
      },
    ];

    tests.forEach(({ name, selections, expected }) => {
      it(name, () => {
        expect(buildWeekdayOfMonthCronFormat(selections)).to.equal(expected);
      });
    });

    it('round-trips through parseWeekdayOfMonth', () => {
      const cron = buildWeekdayOfMonthCronFormat({
        minutes: [0], hours: [3], weekdays: [0, 6], ordinals: [2, LAST],
      });
      expect(parseWeekdayOfMonth(cron)).to.deep.equal({
        weekdays: [0, 6], ordinals: [2, LAST],
      });
    });
  });

  describe('clampOffsetDays', () => {
    const tests = [
      { value: 3, expected: 3 },
      { value: '-2', expected: -2 },
      { value: 2.7, expected: 2 },
      { value: '', expected: 0 },
      { value: '-', expected: 0 },
      { value: MAX_OFFSET_DAYS + 9, expected: MAX_OFFSET_DAYS },
      { value: -MAX_OFFSET_DAYS - 9, expected: -MAX_OFFSET_DAYS },
    ];

    tests.forEach(({ value, expected }) => {
      it(`turns ${JSON.stringify(value)} into ${expected}`, () => {
        expect(clampOffsetDays(value)).to.equal(expected);
      });
    });
  });
});
