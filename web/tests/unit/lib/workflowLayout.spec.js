import { expect } from 'chai';
import { needsAutoLayout, layoutWorkflowNodes } from '@/lib/workflowLayout';

const node = (id, x = 0, y = 0) => ({ id, position_x: x, position_y: y });
const edge = (from, to) => ({ source_node_id: from, destination_node_id: to });

describe('lib/workflowLayout', () => {
  describe('needsAutoLayout', () => {
    const tests = [
      { name: 'null', nodes: null, expected: false },
      { name: 'empty list', nodes: [], expected: false },
      { name: 'all nodes at origin', nodes: [node(1), node(2)], expected: true },
      { name: 'one node positioned', nodes: [node(1), node(2, 10, 0)], expected: false },
      { name: 'only y set', nodes: [node(1, 0, 5)], expected: false },
    ];

    tests.forEach(({ name, nodes, expected }) => {
      it(`returns ${expected} for ${name}`, () => {
        expect(needsAutoLayout(nodes)).to.equal(expected);
      });
    });
  });

  describe('layoutWorkflowNodes', () => {
    it('places a chain into consecutive columns on the same row', () => {
      const positions = layoutWorkflowNodes(
        [node(1), node(2), node(3)],
        [edge(1, 2), edge(2, 3)],
      );

      expect(positions[1].x).to.be.lessThan(positions[2].x);
      expect(positions[2].x).to.be.lessThan(positions[3].x);
      expect(positions[2].x - positions[1].x).to.equal(positions[3].x - positions[2].x);
      expect(positions[1].y).to.equal(positions[2].y);
      expect(positions[2].y).to.equal(positions[3].y);
    });

    it('uses the longest path for the column of a diamond join', () => {
      // 1 -> 2 -> 4, 1 -> 3 -> 4, and a shortcut 1 -> 4
      const positions = layoutWorkflowNodes(
        [node(1), node(2), node(3), node(4)],
        [edge(1, 2), edge(1, 3), edge(2, 4), edge(3, 4), edge(1, 4)],
      );

      expect(positions[2].x).to.equal(positions[3].x);
      expect(positions[4].x).to.be.greaterThan(positions[2].x);
      // siblings in the same column are stacked vertically
      expect(positions[2].y).to.be.lessThan(positions[3].y);
    });

    it('stacks isolated nodes in the first column', () => {
      const positions = layoutWorkflowNodes([node(1), node(2), node(3)], []);

      expect(positions[1].x).to.equal(positions[2].x);
      expect(positions[2].x).to.equal(positions[3].x);
      expect(positions[1].y).to.be.lessThan(positions[2].y);
      expect(positions[2].y).to.be.lessThan(positions[3].y);
    });

    it('ignores edges referencing unknown nodes', () => {
      const positions = layoutWorkflowNodes(
        [node(1), node(2)],
        [edge(1, 99), edge(99, 2), edge(1, 2)],
      );

      expect(Object.keys(positions)).to.have.lengthOf(2);
      expect(positions[2].x).to.be.greaterThan(positions[1].x);
    });

    it('terminates and positions every node when the graph has a cycle', () => {
      const positions = layoutWorkflowNodes(
        [node(1), node(2), node(3)],
        [edge(1, 2), edge(2, 3), edge(3, 2)],
      );

      expect(Object.keys(positions)).to.have.lengthOf(3);
      Object.values(positions).forEach((p) => {
        expect(p.x).to.be.a('number');
        expect(p.y).to.be.a('number');
      });
    });

    it('accepts missing edges', () => {
      const positions = layoutWorkflowNodes([node(1)], undefined);
      expect(positions[1]).to.have.all.keys('x', 'y');
    });
  });
});
