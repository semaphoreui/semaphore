define([
	'app',
	'factories/job'
], function(app) {
	app.registerService('job', ['$http', 'Job', function($http, Job) {
		var self = this;

		self.get = function(playbook, cb) {
			$http.get('/playbook/'+playbook.data._id+'/job/'+job.data._id).success(function(data) {
				self.job = data;
				cb();
			});
		};
	}]);
});