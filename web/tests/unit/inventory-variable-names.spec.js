import { expect } from 'chai';
import variableNames from '@/lib/inventoryVariables';

describe('variableNames', () => {
  [
    ['yaml', 'a: 1\nb: 2', ['a', 'b']],
    ['json', '{"a": 1, "b": {"c": 2}}', ['a', 'b']],
    ['a custom tag', 'a: !vault 1', ['a']],
    ['a blank line in a block scalar', 'banner: |+\n  \nport: 22\n', ['banner', 'port']],
    ['empty text', '', []],
    ['a list', '- 1', []],
  ].forEach(([name, text, names]) => {
    it(`reads ${name}`, () => {
      expect(variableNames(text)).to.deep.equal(names);
    });
  });
});
