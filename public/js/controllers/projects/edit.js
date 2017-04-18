define(function () {
	app.registerController('ProjectEditCtrl', ['$scope', '$http', 'Project', '$state', function ($scope, $http, Project, $state) {
		$scope.projectName = Project.name;
		$scope.alert = Project.alert;

		console.log(Project.name);
		console.log(Project);

		$scope.save = function (name, alert) {
			$http.put(Project.getURL(), { name: name, alert: alert }).success(function () {
				swal('Saved', 'Project settings saved.', 'success');
			}).error(function () {
				swal('Error', 'Project settings were not saved', 'error');
			});
		}

		$scope.deleteProject = function () {
			swal({
				title: 'Delete Project?',
				text: 'All data related to this project will be deleted.',
				type: 'warning',
				showCancelButton: true,
				confirmButtonColor: "#DD6B55",
				confirmButtonText: 'Yes, DELETE'
			}, function () {
				$http.delete(Project.getURL()).success(function () {
					$state.go('dashboard');
				}).error(function () {
					swal('error', 'could not delete project!', 'error');
				});
			});
		}
	}]);
});