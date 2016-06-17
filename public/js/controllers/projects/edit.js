define(function () {
	app.registerController('ProjectEditCtrl', ['$scope', '$http', 'Project', '$state', function ($scope, $http, Project, $state) {
		$scope.projectName = Project.name;

		$scope.save = function (name) {
			$http.put(Project.getURL(), { name: name }).success(function () {
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