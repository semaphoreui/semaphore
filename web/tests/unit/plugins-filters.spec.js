import { expect } from 'chai';
import dayjs from 'dayjs';
import {
  formatDate, formatDate2, formatTime, formatLog, formatMilliseconds,
} from '@/plugins/filters';

const ESC = String.fromCharCode(27);

describe('plugins/filters', () => {
  describe('formatMilliseconds', () => {
    it('returns a dash for empty values', () => {
      expect(formatMilliseconds(null)).to.equal('—');
      expect(formatMilliseconds(undefined)).to.equal('—');
      expect(formatMilliseconds('')).to.equal('—');
    });

    it('humanizes numbers and numeric strings', () => {
      expect(formatMilliseconds(90 * 1000)).to.equal('2 minutes');
      expect(formatMilliseconds(String(2 * 3600 * 1000))).to.equal('2 hours');
      expect(formatMilliseconds(3000)).to.equal('a few seconds');
    });

    it('humanizes a start/end pair', () => {
      expect(formatMilliseconds(['2026-01-01T10:00:00Z', '2026-01-01T10:45:00Z'])).to.equal('an hour');
      expect(formatMilliseconds(['2026-01-01T10:00:00Z', '2026-01-03T10:00:00Z'])).to.equal('2 days');
    });

    it('measures until now when the end is missing', () => {
      const start = dayjs().subtract(5, 'minute').toISOString();
      expect(formatMilliseconds([start, null])).to.equal('5 minutes');
      expect(formatMilliseconds([start, ''])).to.equal('5 minutes');
    });

    it('returns a dash when the start is missing', () => {
      expect(formatMilliseconds([null, '2026-01-01T10:00:00Z'])).to.equal('—');
      expect(formatMilliseconds(['', null])).to.equal('—');
    });

    it('throws on malformed input', () => {
      expect(() => formatMilliseconds([1, 2, 3])).to.throw('invalid value format');
      expect(() => formatMilliseconds({})).to.throw('unsupported value type');
    });
  });

  describe('formatDate', () => {
    it('returns a dash for empty values', () => {
      expect(formatDate(null)).to.equal('—');
      expect(formatDate('')).to.equal('—');
    });

    it('shows relative time with the clock for today', () => {
      const date = dayjs().subtract(2, 'hour');
      const result = formatDate(date.toISOString());
      if (dayjs().isSame(date, 'day')) {
        expect(result).to.equal(`2 hours ago (${date.format('HH:mm')})`);
      } else {
        expect(result).to.equal(date.format('L HH:mm'));
      }
    });

    it('shows the localized date and time for other days', () => {
      const date = dayjs('2020-03-04T05:06:00');
      expect(formatDate(date.toISOString())).to.equal(date.format('L HH:mm'));
    });
  });

  describe('formatDate2 and formatTime', () => {
    it('return a dash for empty values', () => {
      expect(formatDate2(null)).to.equal('—');
      expect(formatTime(null)).to.equal('—');
    });

    it('format a date and a time', () => {
      const date = dayjs('2020-03-04T05:06:07');
      expect(formatDate2(date.toISOString())).to.equal(date.format('LL'));
      expect(formatTime(date.toISOString())).to.equal(date.format('LTS'));
    });
  });

  describe('formatLog', () => {
    it('passes empty values through', () => {
      expect(formatLog('')).to.equal('');
      expect(formatLog(null)).to.equal(null);
    });

    it('converts ANSI colors to styled spans', () => {
      const html = formatLog(`${ESC}[31mfail${ESC}[0m ok`);
      expect(html).to.include('color:rgb(170,0,0)');
      expect(html).to.include('fail');
      expect(html).to.include('ok');
      expect(html).to.not.include(ESC);
    });

    it('escapes HTML in the output', () => {
      expect(formatLog('<b>')).to.equal('&lt;b&gt;');
    });
  });
});
