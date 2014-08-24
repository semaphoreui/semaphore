define(['app'], function (app) {
	app.registerFactory('Playbook', ['$http', function ($http) {
		var Playbook = function (id, cb) {
			if (!id) {
				return;
			}
			
			this.id = id;

			this.get(cb);
		}

		Playbook.prototype.save = function () {
			$http.put('/playbook/'+this.data._id, this.data);
		}

		Playbook.prototype.add = function () {
			$http.post('/playbooks', this.data);
		}

		Playbook.prototype.delete = function () {
			$http.delete('/playbook/'+this.data._id, this.data);	
		}

		Playbook.prototype.get = function (cb) {
			var self = this;

			$http.get('/playbook/'+this.id)
			.success(function (data, status) {
				self.data = data;
				cb();
			})
			.error(function (data, status) {
				cb(data, status);
			})
		}

		return Playbook;
	}])
})