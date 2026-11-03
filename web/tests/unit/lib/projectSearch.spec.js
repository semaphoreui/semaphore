import { expect } from 'chai';
import { filterProjects } from '@/lib/projectSearch';

describe('filterProjects', () => {
  const projects = [
    { id: 1, name: 'Production' },
    { id: 2, name: 'staging' },
    { id: 3, name: 'Customer Prod EU' },
  ];

  it('returns all projects for an empty query', () => {
    expect(filterProjects(projects, '')).to.deep.equal(projects);
    expect(filterProjects(projects, null)).to.deep.equal(projects);
    expect(filterProjects(projects, '   ')).to.deep.equal(projects);
  });

  it('matches a part of the name, ignoring case', () => {
    expect(filterProjects(projects, 'PROD').map((p) => p.id)).to.deep.equal([1, 3]);
    expect(filterProjects(projects, 'stag').map((p) => p.id)).to.deep.equal([2]);
  });

  it('ignores surrounding spaces in the query', () => {
    expect(filterProjects(projects, '  eu ').map((p) => p.id)).to.deep.equal([3]);
  });

  it('ignores diacritics in the name and in the query', () => {
    const accented = [
      { id: 5, name: 'Café Orders' },
      { id: 6, name: 'Zürich Office' },
    ];
    expect(filterProjects(accented, 'cafe').map((p) => p.id)).to.deep.equal([5]);
    expect(filterProjects(accented, 'ZURICH').map((p) => p.id)).to.deep.equal([6]);
    expect(filterProjects(accented, 'zür').map((p) => p.id)).to.deep.equal([6]);
    expect(filterProjects(projects, 'prodúction').map((p) => p.id)).to.deep.equal([1]);
  });

  it('returns an empty list when nothing matches', () => {
    expect(filterProjects(projects, 'dev')).to.deep.equal([]);
  });

  it('handles missing projects and names', () => {
    expect(filterProjects(null, 'x')).to.deep.equal([]);
    expect(filterProjects([{ id: 4 }], 'x')).to.deep.equal([]);
  });
});
