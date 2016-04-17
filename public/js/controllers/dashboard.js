define(['controllers/projects/edit'], function () {
	app.registerController('DashboardCtrl', ['$scope', '$http', '$uibModal', function ($scope, $http, $modal) {
		$scope.projects = [];

		$scope.refresh = function () {
			$http.get('/projects').success(function (projects) {
				$scope.projects = projects;
			});

			$http.get('/events').success(function (events) {
				$scope.events = events;
			});
		}

		$scope.addProject = function () {
			$modal.open({
				templateUrl: '/tpl/projects/add.html'
			}).result.then(function (project) {
				$http.post('/projects', project)
				.success(function () {
					$scope.refresh();
				}).error(function (data, status) {
					swal('Error', 'Could not create project: ' + status, 'error');
				});
			});
		}

		$scope.refresh();
	}]);
});