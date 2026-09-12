import { expect } from 'chai';
import {
  isWeekly,
  isYearly,
  isMonthly,
  isDaily,
  isHourly,
  pruneSelectionsForTiming,
  buildCronFormat,
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
});
