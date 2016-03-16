define([
	'app',
	'jquery',
	'services/job',
	'factories/job'
], function(app, $) {
	app.registerController('PlaybookJobCtrl', ['$scope', '$state', 'job', 'Job', function($scope, $state, job, Job) {

		if(job.data) {
			if (job.data._id.length > 0){
				$scope.job = job;
			}
		}
		else {
			$scope.job = new Job();
		}

		$scope.add = function () {
			$scope.job.add($scope.playbook)
			.success(function (data) {
				$state.transitionTo('playbook.jobs',{
					playbook_id: data.playbook
				});
			})
			.error(function (data) {
			})
		};

		$scope.edit = function () {
			$scope.job.save($scope.playbook)
			.success(function (data) {
				$state.transitionTo('playbook.jobs',{
					playbook_id: $scope.playbook.id
				});
			})
			.error(function (data) {
			})
		}
	}]);
});