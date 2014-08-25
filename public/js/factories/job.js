define(['app'], function (app) {
	app.registerFactory('Job', ['$http', function ($http) {
		var Job = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		Job.prototype.save = function () {
			return $http.put('/job/'+this.data._id, this.data);
		}

		Job.prototype.add = function () {
			return $http.post('/jobs', this.data);
		}

		Job.prototype.delete = function () {
			return $http.delete('/job/'+this.data._id);	
		}

		Job.prototype.get = function () {
			return $http.get('/job/'+this.id);
		}

		return Job;
	}])
})