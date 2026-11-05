import { expect } from 'chai';
import { boolAnswer, normalizeBoolAnswers, surveyDefaults } from '@/lib/surveyVars';

describe('surveyVars', () => {
  describe('boolAnswer', () => {
    [
      [true, true],
      ['true', true],
      [false, false],
      ['false', false],
      ['True', false],
      ['yes', false],
      ['1', false],
      [1, false],
      ['', false],
      [null, false],
      [undefined, false],
    ].forEach(([value, expected]) => {
      it(`${JSON.stringify(value)} -> ${expected}`, () => {
        expect(boolAnswer(value)).to.equal(expected);
      });
    });
  });

  describe('surveyDefaults', () => {
    it('turns a bool default into a boolean and gives every bool an entry', () => {
      expect(surveyDefaults([
        { name: 'on', type: 'bool', default_value: 'true' },
        { name: 'off', type: 'bool', default_value: 'false' },
        { name: 'unset', type: 'bool' },
        { name: 'empty', type: 'bool', default_value: '' },
      ])).to.deep.equal({
        on: true, off: false, unset: false, empty: false,
      });
    });

    it('keeps the defaults of the other types and skips empty ones', () => {
      expect(surveyDefaults([
        { name: 'host', type: '', default_value: 'localhost' },
        { name: 'count', type: 'int', default_value: '3' },
        { name: 'subnets', type: 'select', default_value: ['a'] },
        { name: 'none', type: '', default_value: '' },
        { name: 'missing', type: 'text' },
      ])).to.deep.equal({ host: 'localhost', count: '3', subnets: ['a'] });
    });

    it('accepts no survey', () => {
      expect(surveyDefaults(undefined)).to.deep.equal({});
      expect(surveyDefaults(null)).to.deep.equal({});
    });
  });

  describe('normalizeBoolAnswers', () => {
    const vars = [
      { name: 'a', type: 'bool' },
      { name: 'b', type: 'bool' },
      { name: 'c', type: 'bool' },
      { name: 'd', type: 'bool' },
      { name: 'e', type: 'bool' },
      { name: 'flag', type: '' },
    ];

    it('turns string answers into booleans and leaves other types alone', () => {
      const env = {
        a: 'true', b: 'false', c: true, d: 'yes', flag: 'true',
      };
      expect(normalizeBoolAnswers(vars, env)).to.deep.equal({
        a: true, b: false, c: true, d: false, e: false, flag: 'true',
      });
      expect(env.a).to.equal('true', 'the given object is not changed');
    });

    it('lets an edited answer win over the default when merged first', () => {
      const env = normalizeBoolAnswers(vars, {
        ...surveyDefaults([{ name: 'a', type: 'bool', default_value: 'true' }]),
        a: false,
      });
      expect(env.a).to.equal(false);
    });
  });
});
