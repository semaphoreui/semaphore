import { expect } from 'chai';
import {
  clampZoom, snapToGrid, graphBounds, fitViewport, zoomAt, panBy, screenToCanvas,
  wheelZoomFactor, ZOOM_MIN, ZOOM_MAX,
} from '@/lib/workflowViewport';

describe('lib/workflowViewport', () => {
  describe('clampZoom', () => {
    const tests = [
      { name: 'below min', zoom: 0.01, expected: ZOOM_MIN },
      { name: 'above max', zoom: 10, expected: ZOOM_MAX },
      { name: 'in range', zoom: 1.5, expected: 1.5 },
      { name: 'NaN falls back to min', zoom: NaN, expected: ZOOM_MIN },
    ];
    tests.forEach(({ name, zoom, expected }) => {
      it(name, () => {
        expect(clampZoom(zoom)).to.equal(expected);
      });
    });
  });

  describe('snapToGrid', () => {
    const tests = [
      { value: 0, expected: 0 },
      { value: 9, expected: 0 },
      { value: 10, expected: 20 },
      { value: 31, expected: 40 },
      { value: -11, expected: -20 },
      { value: 123.4, expected: 120 },
    ];
    tests.forEach(({ value, expected }) => {
      it(`${value} -> ${expected}`, () => {
        expect(snapToGrid(value)).to.equal(expected);
      });
    });

    it('accepts a custom grid', () => {
      expect(snapToGrid(13, 8)).to.equal(16);
    });
  });

  describe('graphBounds', () => {
    it('returns null for an empty graph', () => {
      expect(graphBounds([])).to.equal(null);
      expect(graphBounds(null)).to.equal(null);
    });

    it('uses the default node size when none is given', () => {
      const bounds = graphBounds([{ id: 1, x: 100, y: 50 }]);
      expect(bounds).to.deep.equal({
        x: 100, y: 50, width: 220, height: 64,
      });
    });

    it('spans all nodes using their rendered sizes', () => {
      const bounds = graphBounds(
        [{ id: 1, x: 0, y: 0 }, { id: 2, x: 300, y: 200 }],
        { 1: { width: 200, height: 60 }, 2: { width: 100, height: 40 } },
      );
      expect(bounds).to.deep.equal({
        x: 0, y: 0, width: 400, height: 240,
      });
    });
  });

  describe('fitViewport', () => {
    const container = { width: 1000, height: 600 };

    it('centers a small graph at zoom 1 instead of zooming in', () => {
      const bounds = {
        x: 0, y: 0, width: 200, height: 100,
      };
      const vp = fitViewport(bounds, container);
      expect(vp.zoom).to.equal(1);
      expect(vp.x).to.equal(400);
      expect(vp.y).to.equal(250);
    });

    it('zooms out to 0.5 for a graph twice the available width', () => {
      // available width = 1000 - 2*48 = 904 -> bounds 1808 wide fits at 0.5
      const bounds = {
        x: 100, y: 100, width: 1808, height: 100,
      };
      const vp = fitViewport(bounds, container);
      expect(vp.zoom).to.equal(0.5);
      // left edge lands at the padding
      expect(vp.x + bounds.x * vp.zoom).to.equal(48);
      // vertically centered
      expect(vp.y + bounds.y * vp.zoom).to.equal((600 - 50) / 2);
    });

    it('never zooms below the minimum', () => {
      const bounds = {
        x: 0, y: 0, width: 100000, height: 100,
      };
      expect(fitViewport(bounds, container).zoom).to.equal(ZOOM_MIN);
    });

    it('falls back to the identity viewport without bounds or container', () => {
      expect(fitViewport(null, container)).to.deep.equal({ x: 0, y: 0, zoom: 1 });
      expect(fitViewport({
        x: 0, y: 0, width: 1, height: 1,
      }, { width: 0, height: 0 })).to.deep.equal({ x: 0, y: 0, zoom: 1 });
    });
  });

  describe('zoomAt', () => {
    it('keeps the canvas point under the cursor fixed', () => {
      const vp = { x: 100, y: 50, zoom: 1 };
      const point = { x: 400, y: 300 };
      const before = screenToCanvas(vp, point);
      const next = zoomAt(vp, point, 2);
      const after = screenToCanvas(next, point);
      expect(next.zoom).to.equal(2);
      expect(after.x).to.be.closeTo(before.x, 1e-9);
      expect(after.y).to.be.closeTo(before.y, 1e-9);
    });

    it('clamps the requested zoom', () => {
      const next = zoomAt({ x: 0, y: 0, zoom: 1 }, { x: 0, y: 0 }, 100);
      expect(next.zoom).to.equal(ZOOM_MAX);
    });
  });

  it('panBy shifts the translation only', () => {
    expect(panBy({ x: 10, y: 20, zoom: 0.5 }, -5, 15)).to.deep.equal({ x: 5, y: 35, zoom: 0.5 });
  });

  it('wheelZoomFactor zooms out for a positive delta and in for a negative one', () => {
    expect(wheelZoomFactor(100)).to.be.lessThan(1);
    expect(wheelZoomFactor(-100)).to.be.greaterThan(1);
    expect(wheelZoomFactor(0)).to.equal(1);
  });
});
