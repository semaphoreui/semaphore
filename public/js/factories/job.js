define(['app'], function (app) {
	app.registerFactory('Job', ['$http', function ($http) {
		var Job = function (playbook,id,cb) {
			if (!id || !playbook) {
				return;
			}

			this.id = id;
			this.get(playbook,cb);
		}

		Job.prototype.save = function (playbook) {
			return $http.put('/playbook/'+playbook.data._id+'/job/'+this.data._id, this.data);
		}

		Job.prototype.add = function (playbook) {
			return $http.post('/playbook/'+playbook.data._id+'/jobs', this.data);
		}

		Job.prototype.delete = function (playbook) {
			return $http.delete('/playbook/'+playbook.data._id+'/job/'+this.data._id);
		}

		Job.prototype.get = function (playbook,cb) {
			var self = this;

			return $http.get('/playbook/'+playbook.data._id+'/job/'+this.id).success(function (data, status) {
				self.data = data;
				cb();
			})
			.error(function (data, status) {
				console.log(status);
				cb(data, status);
			});
		}

		Job.prototype.run = function (playbook) {
			return $http.post('/playbook/'+playbook.data._id+'/job/'+this.data._id+'/run');
		}

		return Job;
	}])
})