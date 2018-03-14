define([
	'app',
	'factories/job'
], function(app) {
	app.registerService('jobs', ['$http', 'Job', function($http, Job) {
		var self = this;

		self.get = function(playbook, cb) {
			$http.get('/playbook/'+playbook.data._id+'/jobs').then(function(response) {
				self.jobs = [];

				for (var i = 0; i < response.data.length; i++) {
					var job = new Job();
					job.data = response.data[i];
					
					self.jobs.push(job);
				}
				
				cb();
			});
		}
	}]);
});