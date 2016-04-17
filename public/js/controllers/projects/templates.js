define(['controllers/projects/taskRunner'], function () {
	app.registerController('ProjectTemplatesCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$http.get(Project.getURL() + '/keys?type=ssh').success(function (keys) {
			$scope.sshKeys = keys;

			$scope.sshKeysAssoc = {};
			keys.forEach(function (k) {
				$scope.sshKeysAssoc[k.id] = k;
			});
		});
		$http.get(Project.getURL() + '/inventory').success(function (inv) {
			$scope.inventory = inv;

			$scope.inventoryAssoc = {};
			inv.forEach(function (i) {
				$scope.inventoryAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/repositories').success(function (repos) {
			$scope.repos = repos;

			$scope.reposAssoc = {};
			repos.forEach(function (i) {
				$scope.reposAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/environment').success(function (env) {
			$scope.environment = env;

			$scope.environmentAssoc = {};
			env.forEach(function (i) {
				$scope.environmentAssoc[i.id] = i;
			});
		});

		$scope.reload = function () {
			$http.get(Project.getURL() + '/templates').success(function (templates) {
				$scope.templates = templates;
			});
		}

		$scope.remove = function (template) {
			$http.delete(Project.getURL() + '/templates/' + template.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete template..', 'error');
			});
		}

		$scope.add = function () {
			var scope = $rootScope.$new();
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;

			$modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			}).result.then(function (tpl) {
				$http.post(Project.getURL() + '/templates', tpl).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add template:' + status, 'error');
				});
			});
		}

		$scope.run = function (tpl) {
			$modal.open({
				templateUrl: '/tpl/projects/createTaskModal.html',
				controller: 'CreateTaskCtrl',
				resolve: {
					Project: function () {
						return Project;
					},
					Template: function () {
						return tpl;
					}
				}
			})
		}

		$scope.reload();
	}]);
});