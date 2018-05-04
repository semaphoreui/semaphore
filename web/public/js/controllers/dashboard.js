define(['controllers/projects/edit'], function () {
	app.registerController('DashboardCtrl', ['$scope', '$http', '$uibModal', 'SweetAlert', function ($scope, $http, $modal, SweetAlert) {
		$scope.projects = [];

		$scope.refresh = function ($lastEvents=true) {
			$http.get('/projects').then(function (response) {
				$scope.projects = response.data;
			});

			if ($lastEvents == true) {
				$eventsURL = '/events/last';
			} else {
				$eventsURL = '/events';
			}

			$http.get($eventsURL).then(function (response) {
				$scope.events = response.data;
			});
		}

		$scope.addProject = function () {
			$modal.open({
				templateUrl: '/tpl/projects/add.html'
			}).result.then(function (project) {
				$http.post('/projects', project)
				.then(function () {
					$scope.refresh();
				}).catch(function (response) {
					SweetAlert.swal('Error', 'Could not create project: ' + response.status, 'error');
				});
			}, function () {});
		}

		$scope.refresh();
	}]);
});