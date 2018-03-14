define(['app'], function (app) {
	app.registerFactory('Job', ['$http', function ($http) {
		var Job = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
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

		Job.prototype.get = function (playbook) {
			return $http.get('/playbook/'+playbook.data._id+'/job/'+this.id);
		}

		Job.prototype.run = function (playbook) {
			return $http.post('/playbook/'+playbook.data._id+'/job/'+this.data._id+'/run');
		}

		return Job;
	}])
})