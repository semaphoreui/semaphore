import { expect } from 'chai';
import en from '@/lang/en';
import AUDIT_KINDS from '@/lib/auditKinds';
import AuditFilters from '@/components/AuditFilters.vue';

describe('lib/auditKinds', () => {
  it('has an English title for every kind', () => {
    const missing = AUDIT_KINDS.filter((kind) => !en[`audit_kind_${kind.replace(/[./]/g, '_')}`]);
    expect(missing).to.deep.equal([]);
  });
});

describe('AuditFilters.vue kind filter', () => {
  it('offers every kind', () => {
    // Titles come from the global composer, the group headers from $t of the component.
    const items = AuditFilters.computed.kindItems.call({ $t: (k) => k, kindSearch: '' });
    const offered = items.filter((i) => i.value).map((i) => i.value);
    const missing = AUDIT_KINDS.filter((kind) => !offered.includes(kind));
    expect(missing).to.deep.equal([]);
  });
});
