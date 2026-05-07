import { expect } from 'chai';
import en from '@/lang/en';

describe('lang/en.js', () => {
  it('labels the git ref field as Branch / Tag (issue #3812)', () => {
    expect(en.branch).to.equal('Branch / Tag');
  });

  it('uses a branch-or-tag wording for the required-field validation message', () => {
    expect(en.branch_required).to.equal('Branch or tag is required');
  });
});
