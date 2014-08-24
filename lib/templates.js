// By Matej Kramny <matej@matej.me>
// Please leave this comment here.

module.exports = function(app) {
	var self = this;
	self.routes = [];
	self.app = app

	self.route = function (controller) {
		if (!(controller instanceof Array)) {
			controller = [controller];
		}

		for (c in controller) {
			controller[c].unauthorized(self.app, self.add.bind(self));
		}
	}

	self.makeRoute = function(route, view) {
		return {
			route: route,
			view: view
		}
	}

	self.add = function (routes, opts) {
		var args = arguments;

		var prefix = opts ? opts.prefix : null;
		if (!prefix) prefix = '';
		else prefix += '/';

		if (typeof routes === 'string') {
			self.routes.push(self.makeRoute(prefix+routes, prefix+routes));
			return;
		}
		if (Object.prototype.toString.call(routes) == '[object Object]') {
			self.routes.push(routes);
			return;
		}

		for (var i = 0; i < routes.length; i++) {
			var r;
			if (typeof routes[i] == 'string') {
				r = self.makeRoute(prefix+routes[i], prefix+routes[i]);
			} else if (routes[i] instanceof Array) {
				r = self.makeRoute(prefix+routes[i][0], routes[i][1]);
			} else {
				r = routes[i]
			}

			self.routes.push(r);
		}
	}

	self.setup = function () {
		for (var i = 0; i < routes.length; i++) {
			app.get('/view/'+routes[i].route, self.getView.bind(routes[i]));
		}
	}

	self.getView = function (req, res) {
		res.render(this.view);
	}

	return self;
}