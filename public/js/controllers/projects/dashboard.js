define(['controllers/projects/taskRunner'], function () {
	app.registerController('ProjectDashboardCtrl', ['$scope', '$http', 'Project', '$uibModal', '$rootScope', function ($scope, $http, Project, $modal, $rootScope) {
		$http.get(Project.getURL() + '/events').success(function (events) {
			$scope.events = events;
		});

		$scope.reload = function () {
			$http.get(Project.getURL() + '/tasks').success(function (tasks) {
				$scope.tasks = tasks;

				$scope.tasks.forEach(function (t) {
					if (!t.start || !t.end) {
						return;
					}

					// t.duration = moment(t.start).from(moment(t.end), true);
					t.duration = moment(t.start).diff(moment(t.end), 'minutes');
				});
			});
		}
		$scope.reload();

		$scope.openTask = function (task) {
			var scope = $rootScope.$new();
			scope.task = task;
			scope.project = Project;

			$modal.open({
				templateUrl: '/tpl/projects/taskModal.html',
				controller: 'TaskCtrl',
				scope: scope,
				size: 'lg'
			}).result.then(function () {
				$scope.reload();
			});
		}
	}]);
});