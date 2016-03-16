define([
	'app',
	'jquery',
	'services/jobs',
	'factories/job'
], function(app, $) {
	app.registerController('PlaybookJobsCtrl', ['$scope', 'jobs', function($scope, jobs) {
		$scope.jobs = jobs;

		jobs.get($scope.playbook, function () {
		});

		$scope.deleteJob = function (job) {
			job.delete($scope.playbook);

			jobs.get($scope.playbook, function () {
			});
		}

		$scope.runJob = function (job) {
			job.run($scope.playbook);
		}

	}]);
});