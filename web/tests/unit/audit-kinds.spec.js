import { expect } from 'chai';
import en from '@/lang/en';
import AUDIT_KINDS from '@/lib/auditKinds';

describe('lib/auditKinds', () => {
  it('has an English title for every kind', () => {
    const missing = AUDIT_KINDS.filter((kind) => !en[`audit_kind_${kind.replace(/[./]/g, '_')}`]);
    expect(missing).to.deep.equal([]);
  });
});
