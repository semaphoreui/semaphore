define(['controllers/projects/taskRunner'], function () {
	app.registerController('ProjectTemplatesCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$http.get(Project.getURL() + '/keys?type=ssh').success(function (keys) {
			$scope.sshKeys = keys;

			$scope.sshKeysAssoc = {};
			keys.forEach(function (k) {
				if (k.removed) k.name = '[removed] - ' + k.name;
				$scope.sshKeysAssoc[k.id] = k;
			});
		});
		$http.get(Project.getURL() + '/inventory').success(function (inv) {
			$scope.inventory = inv;

			$scope.inventoryAssoc = {};
			inv.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;
				$scope.inventoryAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/repositories').success(function (repos) {
			$scope.repos = repos;

			$scope.reposAssoc = {};
			repos.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;

				$scope.reposAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/environment').success(function (env) {
			$scope.environment = env;

			$scope.environmentAssoc = {};
			env.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;

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
			}).result.then(function (opts) {
				var tpl = opts.template;
				$http.post(Project.getURL() + '/templates', tpl).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add template:' + status, 'error');
				});
			});
		}

		$scope.update = function (template) {
			var scope = $rootScope.$new();
			scope.tpl = template;
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;

			$modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			}).result.then(function (opts) {
				if (opts.remove) {
					return $scope.remove(template);
				}

				var tpl = opts.template;
				$http.put(Project.getURL() + '/templates/' + template.id, tpl).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add template:' + status, 'error');
				});
			}).closed.then(function () {
				$scope.reload();	
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
			}).result.then(function (task) {
				var scope = $rootScope.$new();
				scope.task = task;
				scope.project = Project;

				$modal.open({
					templateUrl: '/tpl/projects/taskModal.html',
					controller: 'TaskCtrl',
					scope: scope,
					size: 'lg'
				});
			})
		}

		$scope.copy = function (template) {
            var tpl = angular.copy(template);
            tpl.id = null;

		    var scope = $rootScope.$new();
			scope.tpl = tpl;
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;

			$modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			}).result.then(function (opts) {
				var tpl = opts.template;
				$http.post(Project.getURL() + '/templates', tpl).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add template:' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});