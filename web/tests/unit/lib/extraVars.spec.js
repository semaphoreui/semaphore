import { expect } from 'chai';
import {
  isPlainObject,
  isJsonSafeValue,
  inferVarType,
  rowToVarValue,
  extraVarsToObject,
  extraVarsToObjectLenient,
  objectToExtraVars,
} from '@/lib/extraVars';

describe('lib/extraVars', () => {
  describe('isPlainObject', () => {
    const tests = [
      { name: 'object literal', value: { a: 1 }, expected: true },
      { name: 'empty object', value: {}, expected: true },
      { name: 'null-prototype object', value: Object.create(null), expected: true },
      { name: 'null', value: null, expected: false },
      { name: 'array', value: [1], expected: false },
      { name: 'string', value: 'x', expected: false },
      { name: 'number', value: 1, expected: false },
      { name: 'Date', value: new Date(), expected: false },
      { name: 'class instance', value: new (class Foo {})(), expected: false },
    ];
    tests.forEach(({ name, value, expected }) => {
      it(`${name} -> ${expected}`, () => {
        expect(isPlainObject(value)).to.equal(expected);
      });
    });
  });

  describe('isJsonSafeValue', () => {
    const tests = [
      { name: 'finite number', value: 42, expected: true },
      { name: 'string', value: 'x', expected: true },
      { name: 'null', value: null, expected: true },
      { name: 'boolean', value: true, expected: true },
      { name: 'NaN', value: NaN, expected: false },
      { name: 'Infinity', value: Infinity, expected: false },
      { name: '-Infinity', value: -Infinity, expected: false },
      { name: 'nested NaN in array', value: [1, [2, NaN]], expected: false },
      { name: 'nested Infinity in object', value: { a: { b: Infinity } }, expected: false },
      { name: 'clean nested structure', value: { a: [1, { b: 'c' }] }, expected: true },
    ];
    tests.forEach(({ name, value, expected }) => {
      it(`${name} -> ${expected}`, () => {
        expect(isJsonSafeValue(value)).to.equal(expected);
      });
    });

    it('does not loop forever on a circular reference', () => {
      const a = { x: 1 };
      a.self = a;
      expect(isJsonSafeValue(a)).to.equal(true);
    });
  });

  describe('inferVarType', () => {
    const tests = [
      { value: [1, 2], expected: 'list' },
      { value: [], expected: 'list' },
      { value: { a: 1 }, expected: 'dict' },
      { value: 3.5, expected: 'number' },
      { value: 'text', expected: 'string' },
      { value: null, expected: 'string' },
      { value: true, expected: 'string' },
      { value: undefined, expected: 'string' },
    ];
    tests.forEach(({ value, expected }) => {
      it(`${JSON.stringify(value)} -> ${expected}`, () => {
        expect(inferVarType(value)).to.equal(expected);
      });
    });
  });

  describe('rowToVarValue', () => {
    it('parses numbers', () => {
      expect(rowToVarValue({ name: 'n', type: 'number', value: '42' })).to.equal(42);
      expect(rowToVarValue({ name: 'n', type: 'number', value: '-1.5' })).to.equal(-1.5);
      expect(rowToVarValue({ name: 'n', type: 'number', value: 7 })).to.equal(7);
    });

    it('rejects empty and non-numeric numbers', () => {
      expect(() => rowToVarValue({ name: 'n', type: 'number', value: '' }))
        .to.throw('Variable "n" must be a number');
      expect(() => rowToVarValue({ name: 'n', type: 'number', value: 'abc' }))
        .to.throw('Variable "n" must be a number');
    });

    it('parses lists', () => {
      expect(rowToVarValue({ name: 'l', type: 'list', value: '["a", 1]' })).to.deep.equal(['a', 1]);
    });

    it('rejects invalid or non-array lists', () => {
      expect(() => rowToVarValue({ name: 'l', type: 'list', value: '[a' }))
        .to.throw('Variable "l" is not a valid list');
      expect(() => rowToVarValue({ name: 'l', type: 'list', value: '{"a":1}' }))
        .to.throw('Variable "l" must be a list');
    });

    it('parses dicts', () => {
      expect(rowToVarValue({ name: 'd', type: 'dict', value: '{"k": "v"}' })).to.deep.equal({ k: 'v' });
    });

    it('rejects invalid, null, array and scalar dicts', () => {
      expect(() => rowToVarValue({ name: 'd', type: 'dict', value: '{' }))
        .to.throw('Variable "d" is not a valid dict');
      ['null', '[1]', '"s"', '5'].forEach((value) => {
        expect(() => rowToVarValue({ name: 'd', type: 'dict', value }), value)
          .to.throw('Variable "d" must be a dict');
      });
    });

    it('passes strings and untyped scalars through unchanged', () => {
      expect(rowToVarValue({ name: 's', type: 'string', value: '42' })).to.equal('42');
      expect(rowToVarValue({ name: 's', type: 'string', value: true })).to.equal(true);
      expect(rowToVarValue({ name: 's', type: undefined, value: 'x' })).to.equal('x');
    });
  });

  describe('extraVarsToObject', () => {
    it('builds an object from typed rows', () => {
      const rows = [
        { name: 'a', type: 'string', value: 'x' },
        { name: 'b', type: 'number', value: '2' },
        { name: 'c', type: 'list', value: '[1]' },
        { name: 'd', type: 'dict', value: '{"k":1}' },
      ];
      expect(extraVarsToObject(rows)).to.deep.equal({
        a: 'x', b: 2, c: [1], d: { k: 1 },
      });
    });

    it('returns an empty object for missing rows', () => {
      expect(extraVarsToObject(null)).to.deep.equal({});
      expect(extraVarsToObject(undefined)).to.deep.equal({});
    });

    it('propagates conversion errors', () => {
      expect(() => extraVarsToObject([{ name: 'b', type: 'number', value: 'x' }]))
        .to.throw('Variable "b" must be a number');
    });

    it('lets a later duplicate name win', () => {
      const rows = [
        { name: 'a', type: 'string', value: 'first' },
        { name: 'a', type: 'string', value: 'second' },
      ];
      expect(extraVarsToObject(rows)).to.deep.equal({ a: 'second' });
    });
  });

  describe('extraVarsToObjectLenient', () => {
    it('keeps the raw text of rows that fail to convert', () => {
      const rows = [
        { name: 'ok', type: 'number', value: '1' },
        { name: 'bad', type: 'list', value: '[unfinished' },
      ];
      expect(extraVarsToObjectLenient(rows)).to.deep.equal({ ok: 1, bad: '[unfinished' });
    });
  });

  describe('objectToExtraVars', () => {
    it('creates typed rows with serialized non-string values', () => {
      expect(objectToExtraVars({
        s: 'text', n: 5, l: [1, 'a'], d: { k: true }, b: false, z: null,
      })).to.deep.equal([
        { name: 's', type: 'string', value: 'text' },
        { name: 'n', type: 'number', value: '5' },
        { name: 'l', type: 'list', value: '[1,"a"]' },
        { name: 'd', type: 'dict', value: '{"k":true}' },
        { name: 'b', type: 'string', value: false },
        { name: 'z', type: 'string', value: null },
      ]);
    });

    it('round-trips through extraVarsToObject', () => {
      const source = {
        s: 'text', n: 5, l: [1, 'a'], d: { k: [1, { x: null }] },
      };
      expect(extraVarsToObject(objectToExtraVars(source))).to.deep.equal(source);
    });
  });
});
